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

package k8s

import (
	"time"
)

func ExampleK8SBlockchain() {
	chain := NewBlockchain(
		4,
		0,
		21000*1000,
		false,
		ImageRepository("quay.io/amis/geth"),
		ImageTag("istanbul_develop"),
		ServiceType("LoadBalancer"),
		Mine(true),
	)
	defer chain.Finalize()

	err := chain.Start(true)
	if err != nil {
		log.Error("Failed to start chain", "err", err)
		return
	}

	<-time.After(20 * time.Second)

	err = chain.Stop(false)
	if err != nil {
		log.Error("Failed to stop chain", "err", err)
		return
	}
}
