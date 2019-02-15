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


package argpool

import (
	"math/big"
	"sort"
)

// simple FIFO queue
type timestamppool struct {
	// Next() iterates circularly so we keep index of current element
	idx int
	// circular list of elements
	storage []*big.Int
	// map for checking whether pool already contains value, for deduplication
	storageMap map[uint64]bool
}

func (p *timestamppool) Add(item uint64) {
	p.storage = append(p.storage, big.NewInt(int64(item)))
	p.storageMap[item] = true
}

func (p *timestamppool) Next() *big.Int {
	if len(p.storage) == 0 {
		return nil
	}
	p.idx = (p.idx + 1) % len(p.storage)
	return p.storage[p.idx]
}

func (p *timestamppool) GetCurrent() *big.Int {
	if len(p.storage) == 0 {
		return nil
	}
	return p.storage[p.idx]
}

func (p *timestamppool) AllPassed() bool {
	if len(p.storage) > 0 && (p.idx+1 == len(p.storage)) {
		return true
	}
	return false
}

func (p *timestamppool) Contains(val uint64) bool {
	_, ok := p.storageMap[val]
	return ok
}

// must only be used for timestamps
func (p *timestamppool) Sort() {
	sort.Slice(p.storage, func(i, j int) bool {
		return p.storage[i].Uint64() < p.storage[j].Uint64()
	})
}

func (p *timestamppool) Size() int {
	return len(p.storage)
}

func GetTimestampPool() *timestamppool {
	return &timestamppool{
		idx:        0,
		storage:    make([]*big.Int, 0),
		storageMap: make(map[uint64]bool),
	}
}
