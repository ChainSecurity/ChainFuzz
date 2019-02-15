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
	"math/rand"
	"reflect"

	"fuzzer/argpool"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

func fillRecursively(T reflect.Type, argPool *argpool.ArgPool) reflect.Value {
	var ret reflect.Value
	ret = reflect.New(T).Elem()

	switch T {
	case Int8Type:
		return reflect.ValueOf(argPool.NextInt8())
	case UInt8Type:
		return reflect.ValueOf(uint8(argPool.NextInt8()))
	case Int16Type:
		return reflect.ValueOf(argPool.NextInt16())
	case UInt16Type:
		return reflect.ValueOf(uint16(argPool.NextInt16()))
	case Int32Type:
		return reflect.ValueOf(argPool.NextInt32())
	case UInt32Type:
		return reflect.ValueOf(uint32(argPool.NextInt32()))
	case Int64Type:
		return reflect.ValueOf(argPool.NextInt64())
	case UInt64Type:
		return reflect.ValueOf(uint64(argPool.NextInt64()))
	case Bytes32Type:
		return reflect.ValueOf(argPool.NextBytes32())
	case AddressType:
		return reflect.ValueOf(argPool.NextAddress())
	case StringType:
		return reflect.ValueOf(argPool.NextString())
	case BoolType:
		x := rand.Int() % 2
		if x == 0 {
			reflect.ValueOf(true)
		}
		return reflect.ValueOf(false)
	case BigIntType:
		return reflect.ValueOf(argPool.NextBigInt())
	}

	if T.Kind() == reflect.Array {
		for i := 0; i < T.Len(); i++ {
			val := fillRecursively(ret.Index(i).Type(), argPool)
			ret.Index(i).Set(val)
		}
		return ret
	}
	if T.Kind() == reflect.Slice {
		len := (rand.Int() & 15) + 1
		for i := 0; i < len; i++ {
			val := fillRecursively(T.Elem(), argPool)
			ret = reflect.Append(ret, val)
		}
		return ret
	}
	panic(fmt.Sprintf("type: %v was not recognized", T))
}

func updateRecursively(val reflect.Value, argPool *argpool.ArgPool) {
	T := val.Type()

	switch T {
	case Int8Type, UInt8Type, BoolType:
		// ignore
		return
	case Int16Type:
		argPool.AddInt16(int16(val.Int()))
		return
	case UInt16Type:
		argPool.AddInt16(int16(val.Uint()))
		return
	case Int32Type:
		argPool.AddInt32(int32(val.Int()))
		return
	case UInt32Type:
		argPool.AddInt32(int32(val.Uint()))
		return
	case Int64Type:
		argPool.AddInt64(val.Int())
		return
	case UInt64Type:
		argPool.AddInt64(int64(val.Uint()))
		return
	case BigIntType:
		argPool.AddBigInt(val.Interface().(*big.Int))
		return
	case AddressType:
		argPool.AddAddress(val.Interface().(common.Address))
		return
	case StringType:
		argPool.AddString(val.String())
		return
	case Bytes32Type:
		argPool.AddBytes32(val.Interface().([32]byte))
		return
	}

	if T.Kind() == reflect.Array || T.Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			updateRecursively(val.Index(i), argPool)
		}
		return
	}
}

// Constructs arguments for function from ABI recursively
func ConstructArgs(args []abi.Argument, argPool *argpool.ArgPool) []interface{} {
	var out []interface{}
	for _, arg := range args {
		val := fillRecursively(arg.Type.Type, argPool)
		out = append(out, val.Interface())
	}
	return out
}

// Updates corresponding pools with values returned from function
// (e.g. if function returns stringpool and int8pool will be updated)
func UpdatePool(input *LastTxInput, argPool *argpool.ArgPool, output []byte) {
	values, err := input.OutArgs.UnpackValues(output)
	if err != nil {
		log.Error(fmt.Sprintf("Error unpacking returned arguments: %+v\n", err))
	}
	for _, value := range values {
		updateRecursively(reflect.ValueOf(value), argPool)
	}
}

func InitArgPool(argPool *argpool.ArgPool, metadata string) {
	// adding dummy bigInt values (powers of 2)
	argPool.AddBigInt(big.NewInt(0))
	argPool.AddInt64(-1)
	for i := int64(1); i < 1E18; i = i * 2 {
		if i <= int64(math.MaxInt32) {
			argPool.AddInt32(int32(i))
		}
		if i <= int64(math.MaxInt16) {
			argPool.AddInt16(int16(i))
		}
		argPool.AddBigInt(big.NewInt(i))
		argPool.AddInt64(i)
	}
	// add random byte32 arrays
	for cnt := 0; cnt < 10; cnt++ {
		byteArr := [32]byte{}
		for i := 0; i < 32; i++ {
			byteArr[i] = byte(rand.Int() % 256)
		}
		argPool.AddBytes32(byteArr)
	}
	// adding dummy string
	argPool.AddString("ChainSecurity")
	// adding dummy byte/bytes
	for i := 0; i < 256; i++ {
		argPool.AddInt8(int8(i))
		argPool.AddInt16(int16(i))
		argPool.AddInt32(int32(i))
		argPool.AddInt64(int64(i))
	}

	accounts := ReadAccounts(metadata)
	// adding dummy addresses
	for _, account := range accounts {
		argPool.AddAddress(account.Address)
		argPool.AddBigInt(account.Amount)
	}
}
