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
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

var argPool *ArgPool

type ArgPool struct {
	Int8Pool    *int8pool
	Int16Pool   *int16pool
	Int32Pool   *int32pool
	Int64Pool   *int64pool
	Bytes32Pool *bytes32pool

	AddressPool   *pool
	BigIntPool    *pool
	StringPool    *stringpool
	TimestampPool *timestamppool
}

func (argPool *ArgPool) AddInt64(item int64) {
	if !argPool.Int64Pool.Contains(item) {
		log.Trace(fmt.Sprintf("adding int64: %+v", item))
		argPool.Int64Pool.Add(item)
	}
}

func (argPool *ArgPool) AddInt32(item int32) {
	if !argPool.Int32Pool.Contains(item) {
		log.Trace(fmt.Sprintf("adding int32: %+v", item))
		argPool.Int32Pool.Add(item)
	}
}

func (argPool *ArgPool) AddInt16(item int16) {
	if !argPool.Int16Pool.Contains(item) {
		log.Trace(fmt.Sprintf("adding int16: %+v", item))
		argPool.Int16Pool.Add(item)
	}
}

func (argPool *ArgPool) AddInt8(item int8) {
	if !argPool.Int8Pool.Contains(item) {
		log.Trace(fmt.Sprintf("adding int8: %+v", item))
		argPool.Int8Pool.Add(item)
	}
}

func (argPool *ArgPool) AddBytes32(item [32]byte) {
	if !argPool.Bytes32Pool.Contains(item) {
		log.Trace(fmt.Sprintf("adding bytes32: %+v", item))
		argPool.Bytes32Pool.Add(item)
	}
}

func (argPool *ArgPool) AddAddress(item common.Address) {
	if !argPool.AddressPool.Contains(item) {
		log.Trace(fmt.Sprintf("adding address: %+v", item))
		argPool.AddressPool.Add(item)
	}
}

func (argPool *ArgPool) AddBigInt(item *big.Int) {
	if !argPool.BigIntPool.Contains(item) {
		log.Trace(fmt.Sprintf("adding bigInt: %+v", item))
		argPool.BigIntPool.Add(item)
	}
}

func (argPool *ArgPool) AddString(item string) {
	if !argPool.StringPool.Contains(item) {
		log.Trace(fmt.Sprintf("adding string: %+v", item))
		argPool.StringPool.Add(item)
	}
}

func (argPool *ArgPool) AddTimestamp(item uint64) {
	if !argPool.TimestampPool.Contains(item) {
		log.Trace(fmt.Sprintf("adding timestamp: %+v", item))
		argPool.TimestampPool.Add(item)
		argPool.TimestampPool.Sort()
	}
}

func (pool *ArgPool) NextInt64() int64 {
	return pool.Int64Pool.Next()
}

func (pool *ArgPool) NextInt32() int32 {
	return pool.Int32Pool.Next()
}

func (pool *ArgPool) NextInt16() int16 {
	return pool.Int16Pool.Next()
}

func (pool *ArgPool) NextInt8() int8 {
	return pool.Int8Pool.Next()
}

func (pool *ArgPool) NextBytes32() [32]byte {
	return pool.Bytes32Pool.Next()
}

func (pool *ArgPool) NextAddress() common.Address {
	return pool.AddressPool.Next().(common.Address)
}

func (pool *ArgPool) NextBigInt() *big.Int {
	return pool.BigIntPool.Next().(*big.Int)
}

func (pool *ArgPool) NextString() string {
	return pool.StringPool.Next()
}

func (pool *ArgPool) NextTimestamp() (*big.Int, bool) {
	if pool.TimestampPool.Size() == 0 {
		return big.NewInt(time.Now().Unix()), false
	}
	return pool.TimestampPool.Next(), pool.TimestampPool.AllPassed()
}

func (pool *ArgPool) CurrentTimestamp() *big.Int {
	if pool.TimestampPool.Size() == 0 {
		return big.NewInt(time.Now().Unix())
	}
	return pool.TimestampPool.GetCurrent()
}

func (pool *ArgPool) GetSizes() map[string]int {
	return map[string]int{
		"int8":      pool.Int8Pool.Size(),
		"int16":     pool.Int16Pool.Size(),
		"int32":     pool.Int32Pool.Size(),
		"int64":     pool.Int64Pool.Size(),
		"string":    pool.StringPool.Size(),
		"address":   pool.AddressPool.Size(),
		"bigInt":    pool.BigIntPool.Size(),
		"timestamp": pool.TimestampPool.Size(),
	}
}

// Singleton, returns same pool for same id
func GetArgPool() *ArgPool {
	if argPool == nil {
		argPool = &ArgPool{
			Int8Pool:      GetInt8Pool(),
			Int16Pool:     GetInt16Pool(),
			Int32Pool:     GetInt32Pool(),
			Int64Pool:     GetInt64Pool(),
			Bytes32Pool:   GetBytes32Pool(),
			AddressPool:   GetPool("Address"),
			BigIntPool:    GetPool("BigInt"),
			StringPool:    GetStringPool(),
			TimestampPool: GetTimestampPool(),
		}
	}
	return argPool
}

func ResetArgPool() {
	argPool = nil
}
