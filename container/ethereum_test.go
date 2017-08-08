// Copyright 2017 AMIS Technologies
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package container

import "testing"

func TestEthereumContainer(t *testing.T) {
	geth := NewEthereum("quay.io/maicoin/ottoman_geth:istanbul_develop", "001")

	err := geth.Start(true)
	if err != nil {
		t.Error(err)
	}

	if !geth.Running() {
		t.Error("geth should be running")
	}

	err = geth.Stop()
	if err != nil {
		t.Error(err)
	}
}
