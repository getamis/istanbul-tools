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

package tests

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/getamis/istanbul-tools/container"
)

var _ = Describe("TSU-04: Non-Byzantine Faulty", func() {
	const (
		numberOfValidators = 4
	)
	var (
		blockchain container.Blockchain
	)

	BeforeEach(func() {
		blockchain = container.NewBlockchain(
			numberOfValidators,
			container.ImageRepository("quay.io/amis/geth"),
			container.ImageTag("istanbul_develop"),
			container.DataDir("/data"),
			container.WebSocket(),
			container.WebSocketAddress("0.0.0.0"),
			container.WebSocketAPI("admin,eth,net,web3,personal,miner"),
			container.WebSocketOrigin("*"),
			container.NAT("any"),
			container.NoDiscover(),
			container.Etherbase("1a9afb711302c5f83b5902843d1c007a1a137632"),
			container.Mine(),
			container.Logging(false),
		)

		Expect(blockchain.Start()).To(BeNil())
	})

	AfterEach(func() {
		blockchain.Stop(true) // This will return container not found error since we stop one
		blockchain.Finalize()
	})

	FIt("TSU-04-01: Stop F validators", func() {

		By("Generating blocks", func() {blockchain.EnsureConsensusWorking(blockchain.Validators(), 10*time.Second)})
		v0 := blockchain.Validators()[0]
		By("Stopping validator 0")
		Expect(v0.Stop()).To(BeNil())

		By("Checking blockchain progress", func() {blockchain.EnsureConsensusWorking(blockchain.Validators()[1:], 20*time.Second)})

	})
})
