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
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const NullAddress = "0000000000000000000000000000000000000000"

type transaction struct {
	Hash             string `json:"hash"`
	AccountNonce     uint64 `json:"nonce"`
	BlockHash        string `json:"blockHash"`
	BlockNumber      int    `json:"blockNumber"`
	TransactionIndex int    `json:"transactionIndex"`
	From             string `json:"from"`
	Recipient        string `json:"to"`
	Amount           string `json:"value"`
	GasLimit         uint64 `json:"gas"`
	Price            string `json:"gasPrice"`
	Payload          string `json:"input"`
}

// removes 0x prefix from hex string and decode payload into bytes
func getPayload(data *string) []byte {
	data_str := *data
	if data_str[1] == 'x' {
		data_str = data_str[2:len(data_str)]
	}
	payload, err := hex.DecodeString(data_str)
	if err != nil {
		panic(fmt.Errorf("error decoding payload: %+v\n", err))
	}
	return payload
}

func isContractCreation(t *transaction) bool {
	if (t.Recipient == "") || (NullAddress == fmt.Sprintf("%x", common.HexToAddress(t.Recipient))) {
		return true
	}
	return false
}

// converts transaction from format that was written as JSON file
// to types.Transaction structure and signs with respective key
func convertAndSign(t *transaction) *types.Transaction {
	// convert amount and gasprice from hex to decimal
	amount, err := strconv.ParseInt(t.Amount, 10, 64)
	if err != nil {
		amount = 0
	}
	price, err := strconv.ParseInt(t.Price, 10, 64)
	if err != nil {
		price = 0
	}
	payload := getPayload(&t.Payload)
	var tx *types.Transaction
	if t.GasLimit > uint64(*maxGasPool) {
		t.GasLimit = uint64(*maxGasPool)
	}
	if isContractCreation(t) {
		tx = types.NewContractCreation(t.AccountNonce, big.NewInt(amount), t.GasLimit, big.NewInt(price), payload)
	} else {
		tx = types.NewTransaction(t.AccountNonce, common.HexToAddress(t.Recipient), big.NewInt(amount), t.GasLimit, big.NewInt(price), payload)
	}
	sender := common.HexToAddress(t.From)
	// mark account as used in order to delete accounts afterwards that weren't
	// used during the deployment
	GetAccountFromAddress(sender).Used = true
	// sign transaction with the sender's key
	signed, err := types.SignTx(tx, types.HomesteadSigner{}, GetKeyFromAddress(sender))
	if err != nil {
		panic(fmt.Errorf("error signing contract: %v\n", err))
	}
	return signed
}

// Reads transactions from json file (that are extracted from Ganache after truffle deploy)
// and returns array of transaction objects
func ReadTransactions(txFile string) []*types.Transaction {
	txJSON, err := os.Open(txFile)
	if err != nil {
		panic(fmt.Errorf("error opening transactions file: %+v\n", err))
	}
	defer txJSON.Close()

	var txs []*types.Transaction
	fileScanner := bufio.NewScanner(txJSON)
	for fileScanner.Scan() {
		t := transaction{}
		err := json.Unmarshal([]byte(fileScanner.Text()), &t)
		if err != nil {
			panic(fmt.Errorf("error unmarshalling transaction: %+v\n", err))
		}
		txs = append(txs, convertAndSign(&t))
	}
	// scanner can't read lines longer than 65536 characters
	if err := fileScanner.Err(); err != nil {
		panic(err)
	}
	return txs
}
