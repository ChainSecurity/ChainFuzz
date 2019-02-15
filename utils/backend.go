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


package utils

import (
	"fmt"
	"math/big"
	"os"

	"fuzzer/argpool"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
)

var (
	maxGasPool = (new(core.GasPool).AddGas(8000000))
	coinBase   = common.HexToAddress(NullAddress)
)

type LastTxInput struct {
	Contract string          `json:"contract"`
	Method   string          `json:"method"`
	Const    bool            `json:"const:"`
	Ether    *big.Int        `json:"ether"`
	Input    *[]interface{}  `json:"arguments"`
	OutArgs  abi.Arguments   `json:"-"`
	Sender   *common.Address `json:"from"`
}

type LastTxResult struct {
	Output           []byte
	RevertAtDepth    int
	AssertionAtDepth int
	Receipt          *types.Receipt
	Overflow         string
	StructLogger     *vm.StructLogger
}

type Backend struct {
	BlockChain        *core.BlockChain
	StateDB           *state.StateDB
	ChainConfig       *params.ChainConfig
	DeployedContracts map[string]*Contract
	// flattened list of deployed contracts for efficient random generation
	ContractsList []string
	// metadata file of truffle project that is being fuzzed
	Metadata string
	// input and output of last transaction
	LastTxIn  *LastTxInput
	LastTxRes LastTxResult

	// instruction indices
	// set of indices for each contract to calculate coverage finally
	OpcodeIndices map[common.Address]map[uint64]bool
	// statistics about execution
	Stats   Stats
	TxCount int
}

func (b *Backend) CommitTransaction(tx *types.Transaction,
	argPool *argpool.ArgPool, options *Options, header *types.Header) (
	error, []*types.Log) {
	b.TxCount = b.TxCount + 1
	b.LastTxRes.StructLogger = nil
	b.LastTxRes.Receipt = nil
	gasPool := *maxGasPool
	err, logs := b.commitTransaction(tx, b.BlockChain, coinBase, &gasPool,
		b.ChainConfig, b.StateDB, header, argPool, options,
	)
	return err, logs
}

func (b *Backend) commitTransaction(
	tx *types.Transaction,
	bc *core.BlockChain,
	coinbase common.Address,
	gp *core.GasPool,
	config *params.ChainConfig,
	state *state.StateDB,
	header *types.Header,
	argPool *argpool.ArgPool,
	options *Options,
) (error, []*types.Log) {
	snap := state.Snapshot()

	logconfig := &vm.LogConfig{
		DisableMemory: false,
		DisableStack:  false,
		// Do not print evm log on console
		Debug: false,
	}

	b.LastTxRes.StructLogger = vm.NewStructLogger(logconfig)
	tracer := b.LastTxRes.StructLogger
	vmConfig := vm.Config{
		Debug:  true,
		Tracer: tracer,
	}
	receipt, _, err := core.ApplyTransaction(config, bc, &coinbase, gp, state,
		header, tx, &header.GasUsed, vmConfig,
	)
	if err != nil {
		state.RevertToSnapshot(snap)
		log.Error(fmt.Sprintf("%v", err))
		return err, nil
	}

	b.processLogs(receipt, tx, argPool, options)

	return nil, receipt.Logs
}

// Returns blockchain, statedb and config for fast fuzzing
// also: - initializes accounts: sets initial balances
// 			 - reads transactions extracted from truffle project
//			 - specified in metadata file and applies to backend
func NewBackend(metadata string, argPool *argpool.ArgPool) *Backend {
	GetABIMap(metadata)
	db, blockchain, err := core.ExpNewCanonical(ethash.NewFullFaker(), 1, true)
	if err != nil {
		panic(fmt.Errorf("error creating blockchain: %v\n", err))
	}
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(db))
	// can't use TestChainConfig, because tracing doesn't work, IsEIP158() messes up stuff
	chainConfig := &params.ChainConfig{
		ChainID:             big.NewInt(1),
		HomesteadBlock:      big.NewInt(0),
		DAOForkBlock:        nil,
		DAOForkSupport:      false,
		EIP150Block:         big.NewInt(0),
		EIP150Hash:          common.Hash{},
		EIP155Block:         big.NewInt(0),
		EIP158Block:         nil, //big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: nil,
		Ethash:              new(params.EthashConfig),
		Clique:              nil,
	}
	// read acccounts from JSON file and set initial balances
	// shadows global variable in this package (accounts.go)
	accounts := ReadAccounts(metadata)
	for _, account := range accounts {
		statedb.SetBalance(account.Address, account.Amount)
	}

	backend := &Backend{
		BlockChain:        blockchain,
		StateDB:           statedb,
		ChainConfig:       chainConfig,
		DeployedContracts: make(map[string]*Contract),
		ContractsList:     make([]string, 0),
		Metadata:          metadata,
		OpcodeIndices:     make(map[common.Address]map[uint64]bool),
	}

	// read transactions json file and apply transactions to backend
	for _, tx := range ReadTransactions(getTxJSONFile(metadata)) {
		backend.CommitTransaction(tx, argPool, &Options{
			UpdateCoverage:        true,
			CheckDeployedContract: true,
			UpdateArgPool:         false,
			ExtractTimestamps:     true,
		}, GetDefaultHeader(backend))
	}
	InitArgPool(argPool, metadata)
	RemoveLibraries(backend)

	for contract, _ := range backend.DeployedContracts {
		for method := range GetContractABI(contract, metadata).Methods {
			backend.DeployedContracts[contract].Methods = append(backend.DeployedContracts[contract].Methods, method)
		}
		if contract != "Migrations" && len(GetContractABI(contract, backend.Metadata).Methods) > 0 {
			backend.ContractsList = append(backend.ContractsList, contract)
		}
	}

	ProcessConfig(metadata, argPool, backend)
	if len(backend.ContractsList) < 1 {
		log.Error(fmt.Sprintf("Truffle deployment script doesn't deploy any contract"))
		os.Exit(0)
	}
	return backend
}
