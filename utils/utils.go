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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type MetadataJSON struct {
	TruffleDir   string `json:"tuffleProjDir"`
	Transactions string `json:"transactions"`
	AccountsFile string `json:"accounts"`
	ConfigFile   string `json:"config"`
}

var metadataJSONMap map[string]*MetadataJSON

// var metadataJSON *MetadataJSON

func getMeta(metadata string) MetadataJSON {
	if metadataJSONMap[metadata] != nil {
		return *metadataJSONMap[metadata]
	}
	content, err := ioutil.ReadFile(metadata)
	if err != nil {
		panic(fmt.Errorf("Error reading metadata: %v\n", metadata))
	}
	metadataJSON := &MetadataJSON{}
	json.Unmarshal([]byte(content), metadataJSON)
	metadataJSONMap[metadata] = metadataJSON
	return *metadataJSON
}

// returns transactions file from metadata json
func getTxJSONFile(metadata string) string {
	meta := getMeta(metadata)
	return meta.Transactions
}

// returns accounts file from metadata json
func getAccountsJSONFile(metadata string) string {
	meta := getMeta(metadata)
	return meta.AccountsFile
}

// reads truffle directory from metadata json
func getTruffleDir(metadata string) string {
	meta := getMeta(metadata)
	return meta.TruffleDir
}

type ContractConfig struct {
	IgnoreAll        bool     `json:"ignore_all"`
	IgnoredFunctions []string `json:"ignore"`
	Timestamps       []uint64 `json:"timestamps"`
}

var fuzzingConfig map[string]ContractConfig

// returns accounts file from metadata json
func GetConfig(metadata string) map[string]ContractConfig {
	if fuzzingConfig != nil {
		return fuzzingConfig
	}

	meta := getMeta(metadata)

	content, err := ioutil.ReadFile(meta.ConfigFile)
	if err != nil {
		panic(fmt.Errorf("Error reading config file: %v\n", meta.ConfigFile))
	}
	fuzzingConfig = make(map[string]ContractConfig)
	json.Unmarshal([]byte(content), &fuzzingConfig)
	return fuzzingConfig
}

func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

// returns file name without extension
func getFilename(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

func init() {
	metadataJSONMap = make(map[string]*MetadataJSON)
}
