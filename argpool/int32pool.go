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

// simple FIFO queue
type int32pool struct {
	// Next() iterates circularly so we keep index of current element
	idx int
	// circular list of elements
	storage []int32
	// map for checking whether pool already contains value, for deduplication
	storageMap map[int32]bool
}

func (p *int32pool) Add(item int32) {
	p.storage = append(p.storage, item)
	p.storageMap[item] = true
}

func (p *int32pool) Next() int32 {
	if len(p.storage) == 0 {
		return 0
	}
	item := p.storage[p.idx]
	p.idx = (p.idx + 1) % len(p.storage)
	return item
}

func (p *int32pool) Contains(val int32) bool {
	_, ok := p.storageMap[val]
	return ok
}

func (p *int32pool) Size() int {
	return len(p.storage)
}

func GetInt32Pool() *int32pool {
	return &int32pool{
		idx:        0,
		storage:    make([]int32, 0),
		storageMap: make(map[int32]bool),
	}
}
