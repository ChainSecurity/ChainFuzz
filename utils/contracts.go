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
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"

	"fuzzer/argpool"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/asm"
	"github.com/ethereum/go-ethereum/log"
)

var (
	ContractHashes    map[string]map[string]string
	DeployedBytecodes map[string]string
	ABIMap            map[string]abi.ABI
	// in method ABI there's no information whether it's payable or not
	// so we keep that info in payable map: payable[contract][method]
	Payable   map[string]map[string]bool
	Libraries map[string]bool
	// swarm hash prefix in bytecode
	bzzr0 = fmt.Sprintf("%x", append([]byte{0xa1, 0x65}, []byte("bzzr0")...))
)

type Contract struct {
	Addresses []common.Address
	Methods   []string
}

// reset global maps, should be used when testing fuzzing
// multiple projects is needed
func ResetContractData() {
	ContractHashes = nil
	DeployedBytecodes = nil
	ABIMap = nil
	Payable = nil
	Libraries = nil
}

func IsPayable(contract, method string) bool {
	return Payable[contract][method]
}

// extracts swarm hash from deployment bytecode
func GetSwarmHash(bytecode string) (string, bool) {
	idx := strings.LastIndex(bytecode, bzzr0)
	if idx == -1 {
		return "", false
	}
	swarmHash := bytecode[idx+len(bzzr0) : idx+len(bzzr0)+64]
	return swarmHash, true
}

func ReadDeployedBytecodes(metadata string) map[string]string {
	if DeployedBytecodes != nil {
		return DeployedBytecodes
	}
	DeployedBytecodes := make(map[string]string)

	truffleDir := getTruffleDir(metadata)
	files, err := ioutil.ReadDir(fmt.Sprintf("%v/build/contracts/", truffleDir))
	if err != nil {
		panic(fmt.Errorf("truffle project folder doesn't exist: %+v\n", err))
	}
	for _, f := range files {
		contractFile := fmt.Sprintf("%v/build/contracts/%v", truffleDir, f.Name())
		contractJSON, _ := ioutil.ReadFile(contractFile)
		var M map[string]string
		json.Unmarshal(contractJSON, &M)
		// trim 0x prefix from deployed code
		DeployedBytecodes[getFilename(f.Name())] = M["deployedBytecode"][2:]
	}
	return DeployedBytecodes
}

// returns map of ContractName -> Swarm hash (of deployment code)
func ReadContractsHashes(metadata string) map[string]string {
	if ContractHashes == nil {
		ContractHashes = make(map[string]map[string]string)
	}
	if ContractHashes[metadata] != nil {
		return ContractHashes[metadata]
	}
	ContractHashes[metadata] = make(map[string]string)

	deployedCodes := ReadDeployedBytecodes(metadata)
	for filename, code := range deployedCodes {
		hash, found := GetSwarmHash(code)
		if !found {
			log.Trace(fmt.Sprintf("swarm hash was not found in contract ABI: %+v",
				filename,
			))
			continue
		}
		ContractHashes[metadata][hash] = filename
	}
	return ContractHashes[metadata]
}

func GetContractNameByHash(hash string, metadata string) string {
	if contractName, found := ReadContractsHashes(metadata)[hash]; found {
		return contractName
	}
	panic(fmt.Sprintf("Contract with hash: %v was not found\n", hash))
}

// returns map of ContractName -> ABI (functs, events ...	)
func GetABIMap(metadata string) map[string]abi.ABI {
	if ABIMap != nil {
		return ABIMap
	}
	type JSONNode struct {
		ContractKind string `json:"contractKind,omitempty"`
		Name         string `json:"name,omitempty"`
	}
	type JSONNodes struct {
		Nodes []JSONNode `json:"nodes"`
	}
	type JSONAST struct {
		ContractName string    `json:"contractName"`
		AST          JSONNodes `json:"ast"`
	}
	type JSONABI struct {
		ABI abi.ABI `json:"abi"`
	}

	truffleDir := getTruffleDir(metadata)
	ABIMap = make(map[string]abi.ABI)
	Payable = make(map[string]map[string]bool)
	Libraries = make(map[string]bool)

	files, err := ioutil.ReadDir(fmt.Sprintf("%v/build/contracts/", truffleDir))
	if err != nil {
		panic(fmt.Errorf("truffle project folder doesn't exist: %+v\n", err))
	}
	for _, f := range files {
		abiFile := fmt.Sprintf("%v/build/contracts/%v", truffleDir, f.Name())
		abiBytes, err := ioutil.ReadFile(abiFile)
		if err != nil {
			panic(fmt.Errorf("Error reading abi file: %+v\n", err))
		}
		dec := json.NewDecoder(strings.NewReader(string(abiBytes)))
		var ast JSONAST
		if err := dec.Decode(&ast); err != nil {
			panic(fmt.Errorf("Error processing contract ast: %+v\n", err))
		}
		for _, node := range ast.AST.Nodes {
			if node.ContractKind == "library" {
				Libraries[node.Name] = true
			}
		}
		// if contract is library ignore it's abi
		if _, ok := Libraries[ast.ContractName]; ok {
			continue
		}
		dec = json.NewDecoder(strings.NewReader(string(abiBytes)))

		var jsonABI JSONABI
		if err := dec.Decode(&jsonABI); err != nil {
			panic(fmt.Errorf("Error processing contract ABI: %+v\n", err))
		}
		evmABI := jsonABI.ABI
		filename := getFilename(f.Name())
		ABIMap[filename] = evmABI
		// Extract payable information about methods
		type method struct {
			Name    string `json:"name"`
			Payable bool   `json:"payable"`
		}
		type methods struct {
			Methods []method `json:"abi"`
		}
		dat := methods{}
		if err := json.Unmarshal(abiBytes, &dat); err != nil {
			panic(err)
		}
		Payable[filename] = make(map[string]bool)
		for _, method := range dat.Methods {
			Payable[filename][method.Name] = method.Payable
		}
	}
	return ABIMap
}

func GetContractABI(contract string, metadata string) abi.ABI {
	evmABI, found := GetABIMap(metadata)[contract]
	if !found {
		panic(fmt.Sprintf("Contract: %v not found", contract))
	}
	return evmABI
}

// returns abi of method
func GetContractMethod(contract string, method string, metadata string) abi.Method {
	evmABI := GetContractABI(contract, metadata)
	if method == "" {
		return evmABI.Constructor
	}
	if abiMethod, found := evmABI.Methods[method]; found {
		return abiMethod
	}
	panic(fmt.Sprintf("Contract: %v, method: %v not found", contract, method))
}

func GetRandomContract(backend *Backend) string {
	size := len(backend.ContractsList)
	return backend.ContractsList[rand.Int()%size]
}

func GetRandomMethod(contract string, backend *Backend) string {
	methods := backend.DeployedContracts[contract].Methods
	return methods[rand.Int()%len(methods)]
}

// returns bytecode for contract method call with specified arguments
func GetCallBytecode(contract string, method string, args []interface{},
	metadata string) []byte {
	evmABI := GetContractABI(contract, metadata)

	input, err := evmABI.Pack(method, args...)
	if err != nil {
		panic(fmt.Errorf("Error generating bytecode for method call: %v, err: %v\n",
			method, err,
		))
	}
	return input
}

func GetOpcodeIndices(metadata string, contract string) []uint64 {
	bytecodes := ReadDeployedBytecodes(metadata)
	bytecode := bytecodes[contract]
	idx := strings.LastIndex(bytecode, bzzr0)
	// last opcode is STOP
	bytecode = bytecode[:idx-1]
	var res []uint64
	script, _ := hex.DecodeString(ReplacePlaceHolders(bytecode))
	it := asm.NewInstructionIterator(script)
	for it.Next() {
		res = append(res, it.PC())
	}
	return res
}

func deleteFromSlice(val string, a []string) []string {
	for i, s := range a {
		if s == val {
			a[i] = a[len(a)-1]
			return a[:len(a)-1]
		}
	}
	return a
}

func IgnoreContract(contract string, backend *Backend) {
	if _, found := backend.DeployedContracts[contract]; found {
		delete(backend.DeployedContracts, contract)
		backend.ContractsList = deleteFromSlice(contract, backend.ContractsList)
	}
}

func IgnoreContractMethod(contract string, method string, backend *Backend) {
	if _, found := backend.DeployedContracts[contract]; found {
		backend.DeployedContracts[contract].Methods = deleteFromSlice(
			method, backend.DeployedContracts[contract].Methods,
		)
	}
}

func ReplacePlaceHolders(bytecode string) string {
	var buffer bytes.Buffer
	for i := 0; i < len(bytecode); i++ {
		if bytecode[i] == '_' {
			i += 39
			buffer.WriteString("0000000000000000000000000000000000000000")
			continue
		}
		buffer.WriteString(string(bytecode[i]))
	}
	return buffer.String()
}

// reads config file and ignores all contract/method during fuzzing according
// to configuration, adds timestamps from config to pool
func ProcessConfig(metadata string, argPool *argpool.ArgPool, backend *Backend) {
	fuzzingConfig := GetConfig(metadata)
	for contract, config := range fuzzingConfig {
		log.Debug(fmt.Sprintf("Ignoring Config for Contract: %v - %v", contract, config))
		for _, timestamp := range config.Timestamps {
			argPool.AddTimestamp(timestamp)
		}
		if config.IgnoreAll {
			IgnoreContract(contract, backend)
			continue
		}
		for _, method := range config.IgnoredFunctions {
			log.Debug(fmt.Sprintf("For Contract %v ignoring %v", contract, method))
			IgnoreContractMethod(contract, method, backend)
		}
	}
}

func RemoveLibraries(backend *Backend) {
	for contract, _ := range Libraries {
		if _, ok := backend.DeployedContracts[contract]; ok {
			log.Debug(fmt.Sprintf("removing library: %v from deployed contracts", contract))
			delete(backend.DeployedContracts, contract)
		}
	}
}
