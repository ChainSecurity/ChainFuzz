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

import "fmt"

type ContractStatsMap map[string][]int
type StatsMap map[string]ContractStatsMap

type Stats struct {
	statsMap         StatsMap
	txCntAfterRevert int
}

func (s *Stats) setIfNil(contract, method string) {
	if s.statsMap == nil {
		s.statsMap = make(StatsMap)
	}
	if s.statsMap[contract] == nil {
		s.statsMap[contract] = make(ContractStatsMap)
	}
	if s.statsMap[contract][method] == nil {
		s.statsMap[contract][method] = []int{0, 0}
	}
}

func (s *Stats) AddFailedTx(contract, method string) {
	s.setIfNil(contract, method)
	s.statsMap[contract][method][0] = s.statsMap[contract][method][0] + 1
	s.statsMap[contract][method][1] = s.statsMap[contract][method][1] + 1
}

func (s *Stats) AddSuccessfulTx(contract, method string) {
	s.setIfNil(contract, method)
	s.statsMap[contract][method][1] = s.statsMap[contract][method][1] + 1
}

func (s *Stats) AddTx(contract, method string, failed bool) {
	s.IncTxCnt()
	if failed {
		s.AddFailedTx(contract, method)
		return
	}
	s.AddSuccessfulTx(contract, method)
}

func (s *Stats) IncTxCnt() {
	s.txCntAfterRevert = s.txCntAfterRevert + 1
}

func (s *Stats) ResetCounter() {
	s.txCntAfterRevert = 0
}

func (s *Stats) GetCount() int {
	return s.txCntAfterRevert
}

func (s *Stats) GetTotalOf(contract, method string) int {
	s.setIfNil(contract, method)
	return s.statsMap[contract][method][1]
}

func (s *Stats) GetStats() map[string]map[string]string {
	res := make(map[string]map[string]string)
	for contract, contractStats := range s.statsMap {
		res[contract] = make(map[string]string)
		for method, pair := range contractStats {
			res[contract][method] = fmt.Sprintf("Failed: %v/%v. (failure rate: %.2f%%)",
				pair[0], pair[1], 100.0*float64(pair[0])/float64(pair[1]),
			)
		}
	}
	return res
}
