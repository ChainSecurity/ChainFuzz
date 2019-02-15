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

	"fuzzer/argpool"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

type Options struct {
	UpdateCoverage        bool
	CheckDeployedContract bool
	UpdateArgPool         bool
	ExtractTimestamps     bool
}

var operators = map[vm.OpCode]func(a, b *big.Int) *big.Int{
	vm.SUB: func(a, b *big.Int) *big.Int {
		res := big.NewInt(0)
		negB := big.NewInt(0)
		return res.Add(a, negB.Neg(b))
	},
	vm.ADD: func(a, b *big.Int) *big.Int {
		res := big.NewInt(0)
		return res.Add(a, b)
	},
	vm.MUL: func(a, b *big.Int) *big.Int {
		res := big.NewInt(0)
		return res.Mul(a, b)
	},
	// ADDMOD, MULMOD are not necessary, they probably should't overflow.
	// as yellowpaper states: All intermediate calculations of this operation
	// (ADDMOD, MULMOD) are not subject to the 2^256 modulo

	vm.EXP: func(a, b *big.Int) *big.Int {
		res := big.NewInt(0)
		return res.Exp(a, b, nil)
	},
}

var (
	earliestTime = uint64(1420070400) //2015.01.01
	latestTime   = uint64(1735689600) // 2025.01.01
)

func getDeployedContractAddress(tx *types.Transaction) common.Address {
	from, err := types.Sender(types.HomesteadSigner{}, tx)
	if err != nil {
		panic(err)
	}
	// address of deployed contract
	return crypto.CreateAddress(from, tx.Nonce())
}

type callStack struct {
	stack []*common.Address
}

func (c *callStack) Push(address *common.Address) {
	c.stack = append(c.stack, address)
}

func (c *callStack) Top() *common.Address {
	return c.stack[len(c.stack)-1]
}

func (c *callStack) Pop() {
	c.stack = c.stack[:len(c.stack)-1]
}

func (b *Backend) processLogs(receipt *types.Receipt, tx *types.Transaction,
	argPool *argpool.ArgPool, options *Options) {

	checkDeployedContract := (options == nil) || options.CheckDeployedContract
	updateCoverage := (options == nil) || options.UpdateCoverage
	// Update coverage if transaction is sent to some contract
	updateCoverage = updateCoverage && (tx.To() != nil)
	updateArgPool := (argPool != nil) && (options != nil) && options.UpdateArgPool

	b.LastTxRes.Output = b.LastTxRes.StructLogger.Output()
	b.LastTxRes.Receipt = receipt

	if updateArgPool && len(b.LastTxRes.Output) > 0 {
		UpdatePool(b.LastTxIn, argPool, b.LastTxRes.Output)
	}

	b.LastTxRes.AssertionAtDepth = -1
	b.LastTxRes.RevertAtDepth = -1
	b.LastTxRes.Overflow = ""
	// there's no INVALID defined in opcodes.go in go-ethereum
	assertOp := vm.OpCode(0xfe)
	structLogs := b.LastTxRes.StructLogger.StructLogs()
	// call stack of contract addresses, top address is currently executing
	// used for coverage calculations, and determining which opcode is from
	// which contract
	callSt := callStack{}
	callSt.Push(tx.To())
	for idx, structLog := range structLogs {

		if options.ExtractTimestamps {
			for _, val := range structLog.Stack {
				if !val.IsUint64() {
					continue
				}
				value := val.Uint64()
				if value > earliestTime && value < latestTime {
					log.Trace(fmt.Sprintf("extracted timestamp: %v", value))
					argPool.AddTimestamp(value)
				}
			}
		}

		// check overflows
		if b.LastTxRes.Overflow == "" {
			if fn, found := operators[structLog.Op]; found {
				st := structLog.Stack
				stLen := len(st)
				operandA := st[stLen-1]
				operandB := st[stLen-2]

				// Out of gas can happen after math opcode, therefore there won't be
				// next log that should include the result
				if idx < len(structLogs) - 1 {
					nextSt := structLogs[idx+1].Stack
					result := nextSt[len(nextSt)-1]
					expected := fn(operandA, operandB)
					if result.Cmp(expected) != 0 {
						b.LastTxRes.Overflow = fmt.Sprintf("(%v %v %v=%v), expected:%v",
							operandA, structLog.Op, operandB, result, expected,
						)
					}
				}
			}
		}

		// check if assertion failure happend in the contract
		if structLog.Op == assertOp {
			if b.LastTxRes.AssertionAtDepth == -1 || b.LastTxRes.AssertionAtDepth > structLog.Depth {
				b.LastTxRes.AssertionAtDepth = structLog.Depth
			}
		}

		// check if REVERT opcode occured in the trace
		if structLog.Op == vm.REVERT {
			if b.LastTxRes.RevertAtDepth == -1 || b.LastTxRes.RevertAtDepth > structLog.Depth {
				b.LastTxRes.RevertAtDepth = structLog.Depth
			}
		}

		// update coverage for initially called contract only
		if updateCoverage {
			if idx > 0 {
				// contract making call to another contract
				if structLog.Depth > structLogs[idx-1].Depth {
					prevStack := structLogs[idx-1].Stack
					// address of callee is second to last in stack of previous structlog
					callee := common.BigToAddress(prevStack[len(prevStack)-2])
					callSt.Push(&callee)
				}
				// return from call
				if structLog.Depth < structLogs[idx-1].Depth {
					callSt.Pop()
				}
			}

			top := *callSt.Top()
			if b.OpcodeIndices[top] == nil {
				b.OpcodeIndices[top] = make(map[uint64]bool)
			}
			b.OpcodeIndices[top][structLog.Pc] = true
		}

		// contract deployment is detected at RETURN opcode:
		// swarm hash is searched in memory
		if !checkDeployedContract || (structLog.Op != vm.RETURN) {
			continue
		}
		// check if contract was deployed
		hash, found := GetSwarmHash(fmt.Sprintf("%x", structLog.Memory))
		if !found {
			continue
		}
		name := GetContractNameByHash(hash, b.Metadata)
		address := getDeployedContractAddress(tx)
		if b.DeployedContracts[name] == nil {
			b.DeployedContracts[name] = &Contract{
				Addresses: make([]common.Address, 0),
				Methods:   make([]string, 0),
			}
		}
		// argPool.AddAddress(address)
		b.DeployedContracts[name].Addresses = append(b.DeployedContracts[name].Addresses, address)
		log.Trace(fmt.Sprintf("Deployed contract: %v with address: %x", name,
			address,
		))
	}
}
