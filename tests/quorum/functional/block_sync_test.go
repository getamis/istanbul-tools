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

package functional

import (
	"context"
	"math/big"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/getamis/istanbul-tools/container"
	"github.com/getamis/istanbul-tools/tests"
)

var _ = Describe("Block synchronization testing", func() {
	const (
		numberOfValidators = 4
	)
	var (
		constellationNetwork container.ConstellationNetwork
		blockchain           container.Blockchain
	)

	BeforeEach(func() {
		constellationNetwork = container.NewDefaultConstellationNetwork(dockerNetwork, numberOfValidators)
		Expect(constellationNetwork.Start()).To(BeNil())
		blockchain = container.NewDefaultQuorumBlockchain(dockerNetwork, constellationNetwork)
		Expect(blockchain.Start(true)).To(BeNil())
	})

	AfterEach(func() {
		blockchain.Stop(true)
		blockchain.Finalize()
		constellationNetwork.Stop()
		constellationNetwork.Finalize()
	})

	Describe("QFS-06: Block synchronization testing", func() {
		const numberOfNodes = 2
		var nodes []container.Ethereum

		BeforeEach(func() {
			var err error

			incubator, ok := blockchain.(container.NodeIncubator)
			Expect(ok).To(BeTrue())

			nodes, err = incubator.CreateNodes(numberOfNodes,
				container.ImageRepository("quay.io/amis/geth"),
				container.ImageTag("istanbul_develop"),
				container.DataDir("/data"),
				container.WebSocket(),
				container.WebSocketAddress("0.0.0.0"),
				container.WebSocketAPI("admin,eth,net,web3,personal,miner"),
				container.WebSocketOrigin("*"),
				container.NAT("any"),
			)

			Expect(err).To(BeNil())

			for _, n := range nodes {
				err = n.Start()
				Expect(err).To(BeNil())
			}
		})

		AfterEach(func() {
			for _, n := range nodes {
				n.Stop()
			}
		})

		It("QFS-06-01: Node connection", func(done Done) {
			By("Connect all nodes to the validators", func() {
				for _, n := range nodes {
					for _, v := range blockchain.Validators() {
						Expect(n.AddPeer(v.NodeAddress())).To(BeNil())
					}
				}
			})

			By("Wait for p2p connection", func() {
				tests.WaitFor(nodes, func(node container.Ethereum, wg *sync.WaitGroup) {
					Expect(node.WaitForPeersConnected(numberOfValidators)).To(BeNil())
					wg.Done()
				})
			})

			close(done)
		}, 15)

		It("QFS-06-02: Node synchronization", func(done Done) {
			const targetBlockHeight = 10

			By("Wait for blocks", func() {
				tests.WaitFor(blockchain.Validators(), func(geth container.Ethereum, wg *sync.WaitGroup) {
					Expect(geth.WaitForBlocks(targetBlockHeight)).To(BeNil())
					wg.Done()
				})
			})

			By("Stop consensus", func() {
				for _, v := range blockchain.Validators() {
					client := v.NewClient()
					Expect(client).NotTo(BeNil())
					err := client.StopMining(context.Background())
					Expect(err).To(BeNil())
					client.Close()
				}
			})

			By("Connect all nodes to the validators", func() {
				for _, n := range nodes {
					for _, v := range blockchain.Validators() {
						Expect(n.AddPeer(v.NodeAddress())).To(BeNil())
					}
				}
			})

			By("Wait for p2p connection", func() {
				tests.WaitFor(nodes, func(node container.Ethereum, wg *sync.WaitGroup) {
					Expect(node.WaitForPeersConnected(numberOfValidators)).To(BeNil())
					wg.Done()
				})
			})

			By("Wait for block synchronization between nodes and validators", func() {
				tests.WaitFor(nodes, func(geth container.Ethereum, wg *sync.WaitGroup) {
					Expect(geth.WaitForBlockHeight(targetBlockHeight)).To(BeNil())
					wg.Done()
				})
			})

			By("Check target block hash of nodes", func() {
				expectedBlock, err := blockchain.Validators()[0].NewClient().BlockByNumber(context.Background(), big.NewInt(targetBlockHeight))
				Expect(err).To(BeNil())
				Expect(expectedBlock).NotTo(BeNil())

				for _, n := range nodes {
					nodeClient := n.NewClient()
					block, err := nodeClient.BlockByNumber(context.Background(), big.NewInt(targetBlockHeight))

					Expect(err).To(BeNil())
					Expect(block).NotTo(BeNil())
					Expect(expectedBlock.Hash()).To(BeEquivalentTo(block.Hash()))
				}
			})

			close(done)
		}, 30)
	})
})
