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
	"crypto/ecdsa"
	"math/big"

	"fuzzer/argpool"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Hint struct {
	Contract string
	Method   string
	Args     []interface{}
	Amount   *big.Int
	Sender   *common.Address
	// generate fallback transaction
	Fallback bool
}

func getRandomAmountFromAddress(address common.Address, argPool *argpool.ArgPool, backend *Backend) *big.Int {
	balance := backend.StateDB.GetBalance(address)
	ret := argPool.NextBigInt()
	for ret.Cmp(balance) >= 1 {
		ret = argPool.NextBigInt()
	}
	return ret
}

func GenTransaction(backend *Backend, argPool *argpool.ArgPool, hint *Hint) *types.Transaction {
	var (
		contract      string
		method        string
		args          []interface{}
		senderAddress *common.Address
		senderKey     *ecdsa.PrivateKey
	)
	if hint != nil && hint.Contract != "" {
		contract = hint.Contract
	} else {
		contract = GetRandomContract(backend)
	}

	if hint != nil && hint.Method != "" {
		method = hint.Method
	} else {
		method = GetRandomMethod(contract, backend)
	}
	if hint.Fallback {
		method = ""
	}

	contractAddress := backend.DeployedContracts[contract].Addresses[0]
	methodABI := GetContractMethod(contract, method, backend.Metadata)
	if hint != nil && hint.Args != nil {
		args = hint.Args
	} else {
		args = ConstructArgs(methodABI.Inputs, argPool)
	}

	input := GetCallBytecode(contract, method, args, backend.Metadata)

	if hint != nil && hint.Sender != nil {
		senderAddress = hint.Sender
		senderKey = GetKeyFromAddress(*senderAddress)
	} else {
		randAccount := GetRandAccount(backend.Metadata)
		senderAddress = &randAccount.Address
		senderKey = randAccount.Key
	}
	// only transfer ether if method is payable
	amount := big.NewInt(0)
	// for each payable function send ether once to cover that case in bytecode
	if hint != nil && hint.Amount != nil {
		amount = hint.Amount
	} else {
		if IsPayable(contract, method) {
			amount = getRandomAmountFromAddress(*senderAddress, argPool, backend)
		}
	}

	// save transaction details in hint object
	hint.Contract = contract
	hint.Method = method
	hint.Amount = amount

	tx := types.NewTransaction(backend.StateDB.GetNonce(*senderAddress),
		contractAddress,
		amount,
		uint64(*maxGasPool),
		big.NewInt(0),
		input,
	)
	signed_tx, _ := types.SignTx(tx, types.HomesteadSigner{}, senderKey)
	backend.LastTxIn = &LastTxInput{
		Contract: contract,
		Method:   method,
		Const:    methodABI.Const,
		Ether:    amount,
		Input:    &args,
		OutArgs:  methodABI.Outputs,
		Sender:   senderAddress,
	}
	return signed_tx
}
