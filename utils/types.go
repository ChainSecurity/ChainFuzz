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
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
)

var (
	Int8Type    = reflect.TypeOf(int8(13))
	UInt8Type   = reflect.TypeOf(uint8(13))
	Int16Type   = reflect.TypeOf(int16(13))
	UInt16Type  = reflect.TypeOf(uint16(13))
	Int32Type   = reflect.TypeOf(int32(13))
	UInt32Type  = reflect.TypeOf(uint32(13))
	Int64Type   = reflect.TypeOf(int64(13))
	UInt64Type  = reflect.TypeOf(uint64(13))
	Bytes32Type = reflect.TypeOf([32]byte{})
	AddressType = reflect.TypeOf(common.Address{})
	BigIntType  = reflect.TypeOf((*big.Int)(nil))
	BoolType    = reflect.TypeOf(true)
	StringType  = reflect.TypeOf("Expecto Patronum")
)
