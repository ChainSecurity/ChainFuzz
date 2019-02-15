/***
*  
*  ChainSecurity ChainFuzz - a fast ethereum transaction fuzzer
*  Copyright (C) 2019 ChainSecurity AG
*  
*  This program is free software: you can redistribute it and/or modify
*  it under the terms of the GNU Affero General Public License as published by
*  the Free Software Foundation, either version 3 of the License, or
*  (at your option) any later version.
*  
*  This program is distributed in the hope that it will be useful,
*  but WITHOUT ANY WARRANTY; without even the implied warranty of
*  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
*  GNU Affero General Public License for more details.
*  
*  You should have received a copy of the GNU Affero General Public License
*  along with this program.  If not, see <https://www.gnu.org/licenses/>.
*
***/


package main

import (
	"fmt"
	"math/big"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
	"strings"

	"fuzzer/argpool"
	"fuzzer/utils"

	"github.com/ethereum/go-ethereum/log"
	"gopkg.in/urfave/cli.v1"
)

const (
	LatestStateRootKey = "evm:LatestStateRootKey"
)

var (
	app          *cli.App
	logLevelFlag = cli.IntFlag{
		Name:  "loglevel",
		Usage: "log level to emit to the screen",
		Value: int(log.LvlInfo),
	}
	metadataFileFlag = cli.StringFlag{
		Name:  "metadata",
		Usage: "metadata file of truffle project to fuzz (generated from extract.sh script)",
		Value: "",
	}
	contractFlag = cli.StringFlag{
		Name:  "contract",
		Usage: "name of contract to fuzz",
		Value: "",
	}
	limitFlag = cli.IntFlag{
		Name:  "limit",
		Usage: "stop fuzzing after specified number of transactions",
		Value: 10000,
	}
	// optimizations mode is encoded in bits (similar to unix file permissions)
	// see OptMode struct
	optimizationsFlag = cli.IntFlag{
		Name:  "o",
		Usage: "bitset of optimizations mode (check struct OptMode for more)",
		Value: 0,
	}
	interactiveModeFlag = cli.BoolFlag{
		Name:  "it",
		Usage: "Interactive mode for guiding fuzzer/checking state",
		// Value: false,
	}
	// number of accounts to be used additionally to what is was during deployment
	accountsFlag = cli.IntFlag{
		Name:  "a",
		Usage: "number of accounts to use during fuzzing (additionally to  what is was during deployment)",
		Value: 1,
	}
	cpuProfileFlag = cli.StringFlag{
		Name: "cpuprofile",
		// example --cpuprofile cpu.prof
		Usage: "flag for profiling cpu, specify filename to save profile",
		Value: "",
	}
	memProfileFlag = cli.StringFlag{
		Name:  "memprofile",
		Usage: "flag for profiling memory, specify filename to save profile",
		Value: "",
	}
)

type Flags struct {
	Metadata    string
	Contract    string
	Limit       int
	Interactive bool
	OptMode     *utils.OptMode
	Accounts    int
	LogLevel    int
}

func init() {
	app = cli.NewApp()
	app.Name = "Fuzzer"
	app.Author = "ChainSecurity"
	app.Email = "contact@chainsecurity.com"
	app.Version = "0.1"
	app.Usage = "Fuzzing Ethereum smart contracts"
	app.Flags = []cli.Flag{
		metadataFileFlag,
		contractFlag,
		limitFlag,
		optimizationsFlag,
		interactiveModeFlag,
		logLevelFlag,
		cpuProfileFlag,
		memProfileFlag,
	}
	app.Action = run
}

func getCLFlags(ctx *cli.Context) *Flags {
	metadata := ctx.GlobalString(metadataFileFlag.Name)
	if metadata == "" {
		log.Error("please provide metadata file with --metadata option")
		os.Exit(1)
	}
	contract := ctx.GlobalString(contractFlag.Name)
	if contract == "" {
		log.Warn("Contract was not provided (you can use --contract option)")
	}
	limit := ctx.GlobalInt(limitFlag.Name)
	optimizations := ctx.GlobalInt(optimizationsFlag.Name)
	optMode := &utils.OptMode{}
	optMode.SetFlag(optimizations)
	accounts := ctx.GlobalInt(accountsFlag.Name)

	return &Flags{
		Metadata:    metadata,
		Contract:    contract,
		Limit:       limit,
		Interactive: ctx.GlobalBool(interactiveModeFlag.Name),
		OptMode:     optMode,
		Accounts:    accounts,
		LogLevel:    ctx.Int(logLevelFlag.Name),
	}
}

func getCoverage(metadata string, contract string, backend *utils.Backend) (int, int) {
	indices := utils.GetOpcodeIndices(metadata, contract)
	contractAddress := backend.DeployedContracts[contract].Addresses[0]
	coveredIndices := backend.OpcodeIndices[contractAddress]
	return len(coveredIndices), len(indices)
}

// resets all global variables that are used for caching
// only useful when testing several projects at the same time
func Reset() {
	utils.ResetContractData()
	argpool.ResetArgPool()
}

func fuzz(ctx *cli.Context) {
	flags := getCLFlags(ctx)
	argPool := argpool.GetArgPool()
	backend := utils.NewBackend(flags.Metadata, argPool)
	result := make(utils.ResultsMap)
	utils.SnapshotBackend(backend)

	// start timer
	start := time.Now()

	options := &utils.Options{
		UpdateCoverage:        true,
		CheckDeployedContract: false,
		UpdateArgPool:         true,
	}

	// initial transactions to cover corner cases:
	// - fallbacks for all contracts
	// - paying non-payable methods
	for contract, _ := range backend.DeployedContracts {
		if contract == "Migrations" {
			continue
		}
		hint := &utils.Hint{Contract: contract, Fallback: true}
		utils.Rec(backend, argPool, hint, options, result, flags.OptMode, 0)
		for _, method := range backend.DeployedContracts[contract].Methods {
			// send ether to all un-payable function except fuzz_always_true
			if !utils.IsPayable(contract, method) && strings.Index(method, "fuzz_always_true") != 0 {
				hint = &utils.Hint{Contract: contract, Method: method, Amount: big.NewInt(1)}
				utils.Rec(backend, argPool, hint, options, result, flags.OptMode, 0)
			}
		}
	}

	backend.TxCount = 0
	// increase limit by number of deployment transactions
	flags.Limit = flags.Limit + backend.TxCount
	for i := backend.TxCount; i < flags.Limit; i++ {
		// ADVICED NOT TO SPECIFY HINT FOR ALL TRANSACTIONS (in Hint obj)
		// even when you only want to test specific contract extracted values
		// from other contracts are important for fuzzing main contract
		hint := &utils.Hint{}
		terminate := utils.Rec(backend, argPool, hint, options, result, flags.OptMode, 0)
		if terminate || backend.TxCount > flags.Limit {
			break
		}
		if flags.OptMode.SnapshotsEnabled && (i&8192 != 0) {
			utils.RevertBackend(backend)
		}

		if flags.LogLevel <= int(log.LvlInfo) {
			fmt.Printf("\rTransactions:  %v/%v, %v%%", backend.TxCount, flags.Limit, 100.0*backend.TxCount/flags.Limit)
		}
	}
	fmt.Println()

	// stop timer, and check coverage
	elapsed := time.Since(start)
	log.Info("Calculating coverage. (needs to read and analyse bytecode opcodes)")
	for contract, _ := range backend.DeployedContracts {
		if result[contract] == nil {
			continue
		}

		covered, total := getCoverage(flags.Metadata, contract, backend)
		s := fmt.Sprintf("%v/%v, %v%%", covered, total, 100.0*covered/total)
		result[contract]["covered"] = s
	}

	log.Info(fmt.Sprintf("Fuzzing result: %+v", utils.PrettyPrint(result)))
	log.Trace(fmt.Sprintf("Arg pool sizes: %+v", utils.PrettyPrint(argPool.GetSizes())))
	if flags.OptMode.GenStatistics {
		log.Info(fmt.Sprintf("Stats: %+v", utils.PrettyPrint(backend.Stats.GetStats())))
	}

	log.Info(fmt.Sprintf("fuzzed: %v tx in %.2fsec. (rate: %.2f tx/s)",
		backend.TxCount, elapsed.Seconds(), float64(backend.TxCount)/elapsed.Seconds(),
	))
}

func run(ctx *cli.Context) error {
	cpuprofile := ctx.GlobalString(cpuProfileFlag.Name)
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Error(fmt.Sprintf("could not create CPU profile: ", err))
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Error(fmt.Sprintf("could not start CPU profile: ", err))
		}
		defer pprof.StopCPUProfile()
	}

	log.Root().SetHandler(log.LvlFilterHandler(
		log.Lvl(ctx.Int(logLevelFlag.Name)),
		log.StreamHandler(os.Stderr, log.TerminalFormat(true)),
	))
	fuzz(ctx)

	memprofile := ctx.GlobalString(memProfileFlag.Name)
	if memprofile != "" {
		f, err := os.Create(memprofile)
		if err != nil {
			log.Error(fmt.Sprintf("could not create memory profile: ", err))
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Error(fmt.Sprintf("could not write memory profile: ", err))
		}
		f.Close()
	}
	return nil
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
