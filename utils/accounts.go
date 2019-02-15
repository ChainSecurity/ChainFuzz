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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// non exported structure used only for parsing accounts json
type accountJSON struct {
	Key    string   `json:"key"`
	Amount *big.Int `json:"amount"`
}

// non exported structure used only for parsing accounts json
type profileJSON struct {
	Accounts []accountJSON `json:"accounts"`
}

type Account struct {
	Key     *ecdsa.PrivateKey
	Address common.Address
	Amount  *big.Int
	// field to mark if account was used during deployment
	Used bool
}

var (
	accounts     []Account
	keyToAddress map[ecdsa.PrivateKey]common.Address
	addressToKey map[common.Address]*ecdsa.PrivateKey
)

func GetAddressFromKey(key ecdsa.PrivateKey) common.Address {
	if address, found := keyToAddress[key]; found {
		return address
	}
	panic(fmt.Sprintf("key: %v was not found in accounts", key))
}

func GetKeyFromAddress(address common.Address) *ecdsa.PrivateKey {
	if key, found := addressToKey[address]; found {
		return key
	}
	panic(fmt.Sprintf("address: %v was not found in accounts", address))
}

func GetAccountFromAddress(address common.Address) *Account {
	for i := 0; i < len(accounts); i++ {
		if accounts[i].Address == address {
			return &accounts[i]
		}
	}
	panic(fmt.Sprintf("account wasn't found for address: %v", address))
}

// singleton pattern
// reads accounts from JSON file and returns array of accounts
// containing Key, Address and initial Balance for account
func ReadAccounts(metadata string) []Account {
	if accounts != nil {
		return accounts
	}
	f := getAccountsJSONFile(metadata)
	accountsJSON, err := ioutil.ReadFile(f)
	if err != nil {
		panic(fmt.Errorf("error reading accounts file: %v %+v\n", f, err))
	}
	var p profileJSON
	err = json.Unmarshal(accountsJSON, &p)
	if err != nil {
		panic(err)
	}
	accounts = make([]Account, 0)
	keyToAddress = make(map[ecdsa.PrivateKey]common.Address)
	addressToKey = make(map[common.Address]*ecdsa.PrivateKey)
	for _, account := range p.Accounts {
		key, err := crypto.HexToECDSA(account.Key)
		if err != nil {
			panic(fmt.Errorf("Incorrect hex key: %+v", err))
		}
		acc := Account{
			Key:     key,
			Address: crypto.PubkeyToAddress(key.PublicKey),
			Amount:  account.Amount,
		}
		accounts = append(accounts, acc)
		addressToKey[acc.Address] = acc.Key
		keyToAddress[*acc.Key] = acc.Address
	}
	return accounts
}

func GetRandAccount(metadata string) Account {
	n := len(ReadAccounts(metadata))
	return ReadAccounts(metadata)[rand.Intn(n)]
}

func init() {
}
