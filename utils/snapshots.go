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

	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/log"
)

var backendSnapshot *state.StateDB

func SnapshotBackend(backend *Backend) {
	backendSnapshot = backend.StateDB.Copy()
	log.Trace(fmt.Sprintf("STATE: new snapshot version has been created"))
}

func RevertBackend(backend *Backend) {
	*backend.StateDB = *backendSnapshot
	log.Trace(fmt.Sprintf("STATE: state has been reverted"))
}
