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
	"fmt"
)

var pools map[string]*pool

// simple FIFO queue
type pool struct {
	// Next() iterates circularly so we keep index of current element
	idx int
	// circular list of elements
	storage []interface{}
	// map for checking whether pool already contains value, for deduplication
	storageMap map[string]bool
}

func (p *pool) Add(item interface{}) {
	p.storage = append(p.storage, item)
	p.storageMap[fmt.Sprintf("%v", item)] = true
}

func (p *pool) Next() interface{} {
	if len(p.storage) == 0 {
		return nil
	}
	item := p.storage[p.idx]
	p.idx = (p.idx + 1) % len(p.storage)
	return item
}

func (p *pool) Contains(val interface{}) bool {
	_, ok := p.storageMap[fmt.Sprintf("%v", val)]
	return ok
}

func (p *pool) Size() int {
	return len(p.storage)
}

func GetPool(id string) *pool {
	if pools[id] == nil {
		pools[id] = &pool{
			idx:        0,
			storage:    make([]interface{}, 0),
			storageMap: make(map[string]bool),
		}
	}
	return pools[id]
}

func init() {
	pools = make(map[string]*pool)
}
