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
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/getamis/istanbul-tools/container"
)

var _ = Describe("TFS-01: General consensus", func() {
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

	It("TFS-01-03: Peer connection", func() {

		By("Check peer count")
		errc := make(chan error, numberOfValidators)
		for _, v := range blockchain.Validators() {
			go func(v container.Ethereum) {
				c := v.NewIstanbulClient()
				ticker := time.NewTicker(time.Millisecond * 100)
				timeout := time.NewTimer(time.Second * 10)
				expPeerCnt := numberOfValidators - 1
				for {
					select {
					case <-ticker.C:
						peers, err := c.AdminPeers(context.Background())
						Expect(err).To(BeNil())
						if len(peers) != expPeerCnt {
							continue
						} else {
							errc <- nil
							return
						}
					case <-timeout.C:
						errc <- errors.New("Check peer count timeout.")
						return
					}
				}
			}(v)
		}

		for i := 0; i < numberOfValidators; i++ {
			err := <-errc
			Expect(err).To(BeNil())
		}
	})
})
