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
	"context"
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

	It("TSU-04-01: Stop F validators", func(done Done) {

		By("Generating blocks")
		v0 := blockchain.Validators()[0]
		c0 := v0.NewIstanbulClient()
		ticker := time.NewTicker(time.Millisecond * 100)
		for _ = range ticker.C {
			n, e := c0.BlockNumber(context.Background())
			Expect(e).To(BeNil())
			// Check if new blocks are getting generated
			if n.Int64() > 1 {
				ticker.Stop()
				break
			}
		}
		By("Stopping validator 0")
		e := v0.Stop()
		Expect(e).To(BeNil())

		ticker = time.NewTicker(time.Millisecond * 100)
		for _ = range ticker.C {
			e := v0.Stop()
			// Wait for e to be non-nil to make sure the container is down
			if e != nil {
				ticker.Stop()
				break
			}
		}

		By("Checking blockchain progress")
		v1 := blockchain.Validators()[1]
		c1 := v1.NewIstanbulClient()
		n1, e := c1.BlockNumber(context.Background())
		Expect(e).To(BeNil())
		ticker = time.NewTicker(time.Millisecond * 100)
		for _ = range ticker.C {
			newN1, e := c1.BlockNumber(context.Background())
			Expect(e).To(BeNil())
			if newN1.Int64() > n1.Int64() {
				ticker.Stop()
				break
			}
		}

		close(done)
	}, 80)
})
