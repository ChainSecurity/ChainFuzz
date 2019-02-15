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
	"math"
	"math/big"
	"strings"
	"time"

    "encoding/json"
    "io/ioutil"

	"fuzzer/argpool"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

type ResultsMap map[string]map[string]string

// Bitset of optimizations optimizations mode
// bits from least significant bit to most
type OptMode struct {
	// if transaction fails retry it with half the ether
	RetryHalfEther bool
	// retry failed transaction with different sender
	RetryDiffSender bool
	// snapshot state and if "too many transactions fail" revert the state
	SnapshotsEnabled bool
	// process statistics, which method was called how many times and rate of failure
	GenStatistics bool
}

func (o *OptMode) SetFlag(optFlag int) {
	if optFlag&1 != 0 {
		o.RetryHalfEther = true
	}
	if optFlag&2 != 0 {
		o.RetryDiffSender = true
	}
	if optFlag&4 != 0 {
		o.SnapshotsEnabled = true
	}
	if optFlag&8 != 0 {
		o.GenStatistics = true
	}
}

func GenTimestamp(backend *Backend, argPool *argpool.ArgPool) *big.Int {
	// submit 2048 transactions for each counter
	if backend.Stats.GetCount() < 2048 {
		return argPool.CurrentTimestamp()
	}
	timestamp, shouldRevert := argPool.NextTimestamp()
	backend.Stats.ResetCounter()
	if shouldRevert {
		RevertBackend(backend)
	}
	return timestamp
}

func SaveJson(filename string, backend *Backend){
    jsonOut, _ := json.MarshalIndent(backend.LastTxRes.StructLogger.StructLogs(),"","  ")
    _ = ioutil.WriteFile(filename, jsonOut, 0644)
    log.Debug(fmt.Sprintf("Saved transaction to %v.", filename))
}


func Rec(backend *Backend, argPool *argpool.ArgPool, hint *Hint,
		options *Options, result ResultsMap, optMode *OptMode, depth int) bool {

	header := GetDefaultHeader(backend)
	header.Time = GenTimestamp(backend, argPool)
	tx := GenTransaction(backend, argPool, hint)
	desc := backend.LastTxIn
	if result[desc.Contract] == nil {
		result[desc.Contract] = make(map[string]string)
	}
	log.Debug(fmt.Sprintf(fmt.Sprintf("%v. Fuzzing:\t%+v", backend.TxCount+1, PrettyPrint(desc))))
	backend.CommitTransaction(tx, argPool, options, header)

	log.Debug(fmt.Sprintf("output: %+v\n", backend.LastTxRes.StructLogger.Output()))

	reverted := backend.LastTxRes.RevertAtDepth == 1
	// If overflow happens and transaction is not reverted
	if backend.LastTxRes.Overflow != "" && !reverted {
		log.Debug("Overflow detected")
        log.Debug(backend.LastTxRes.Overflow)
        SaveJson(fmt.Sprintf("/tmp/overflow_%v.json", backend.TxCount), backend)
		s := fmt.Sprintf("%v: Overflow", desc.Method)
		result[desc.Contract][s] = backend.LastTxRes.Overflow
	}

	if backend.LastTxRes.AssertionAtDepth != -1 {
		log.Debug(fmt.Sprintf("Assertion failure.\ttook: %v transactions", backend.TxCount))
		s := fmt.Sprintf("%v: %v", desc.Method, "AssertionFailure")
        SaveJson(fmt.Sprintf("/tmp/assertion_%v.json", backend.TxCount), backend)
		result[desc.Contract][s] = ""
	}

	if strings.Index(desc.Method, "fuzz_always_true") == 0 {

		if !reverted && len(backend.LastTxRes.StructLogger.Output()) == 32 && backend.LastTxRes.StructLogger.Output()[31] != 1 {
			log.Debug(fmt.Sprintf("property violation, not always true.\ttook: %v transactions", backend.TxCount))
			s := fmt.Sprintf("%v: %v", desc.Method, "Property violation")
            SaveJson("/tmp/violation.json", backend)

			result[desc.Contract][s] = ""
			return true
		}

		// stop fuzzing once a fuzz_always_true function reverts
		if reverted {
			log.Debug(fmt.Sprintf("revert in fuzz function.\ttook: %v transactions", backend.TxCount))
			s := fmt.Sprintf("%v: %v", desc.Method, "Revert in fuzz function")
            SaveJson("/tmp/reverted_fuzz.json", backend)

			result[desc.Contract][s] = ""
			return true
		}
	}

	if optMode.GenStatistics {
		failed := backend.LastTxRes.RevertAtDepth == 1
		backend.Stats.AddTx(backend.LastTxIn.Contract, backend.LastTxIn.Method, failed)
	} else {
		// update transactions counter even if statistics option is not used
		backend.Stats.IncTxCnt()
	}

	// Heuristics for reverted transactions
	// if transaction was not reverted return
	if backend.LastTxRes.RevertAtDepth != 1 {
		return false
	}

	if optMode.RetryHalfEther {
		// half the value of ether for transaction
		if tx.Value().Sign() == 1 {
			// divide by 2 (bitwise right shift by 1)
			hint.Amount.Rsh(hint.Amount, 1)
			hint.Sender = backend.LastTxIn.Sender
			Rec(backend, argPool, hint, options, result, optMode, depth+1)
		}
	}

	if optMode.RetryDiffSender {
		if depth < 4 {
			log.Trace("Retrying transaction with different sender")
			hint.Sender = nil
			Rec(backend, argPool, hint, options, result, optMode, depth+1)
		}
	}
	return false
}

var defaultHeader *types.Header

func GetDefaultHeader(b *Backend) *types.Header {
	if defaultHeader == nil {
		defaultHeader = &types.Header{
			Coinbase:   coinBase,
			ParentHash: b.BlockChain.CurrentBlock().Hash(),
			Number:     big.NewInt(1),
			GasLimit:   math.MaxUint64,
			Difficulty: big.NewInt(int64(1)),
			Extra:      nil,
			Time:       big.NewInt(time.Now().Unix()),
		}
	}
	return defaultHeader
}
