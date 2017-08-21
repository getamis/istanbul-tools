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
	"math/big"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/getamis/istanbul-tools/container"
)

var _ = Describe("4 validators Istanbul", func() {
	const (
		numberOfValidators = 4
	)
	var (
		blockchain container.Blockchain
	)

	BeforeSuite(func() {
		blockchain = container.NewBlockchain(
			numberOfValidators,
			container.ImageRepository("quay.io/amis/geth"),
			container.ImageTag("istanbul_develop"),
			container.DataDir("/data"),
			container.WebSocket(),
			container.WebSocketAddress("0.0.0.0"),
			container.WebSocketAPI("admin,eth,net,web3,personal,miner,istanbul"),
			container.WebSocketOrigin("*"),
			container.NAT("any"),
			container.NoDiscover(),
			container.Etherbase("1a9afb711302c5f83b5902843d1c007a1a137632"),
			container.Mine(),
			container.Logging(true),
		)

		Expect(blockchain.Start()).To(BeNil())
	})

	AfterSuite(func() {
		Expect(blockchain.Stop()).To(BeNil())
		blockchain.Finalize()
	})

	It("Blockchain creation", func() {
		for _, geth := range blockchain.Validators() {
			client := geth.NewClient()
			Expect(client).NotTo(BeNil())

			block, err := client.BlockByNumber(context.Background(), big.NewInt(0))
			Expect(err).To(BeNil())
			Expect(block).NotTo(BeNil())
		}

		By("Ensure that consensus is working in 30 seconds", func() { ensureConsensusWorking(blockchain.Validators(), 30*time.Second) })
	})
})

func ensureConsensusWorking(geths []container.Ethereum, t time.Duration) {
	errCh := make(chan error, len(geths))
	quitCh := make(chan struct{}, len(geths))
	for _, geth := range geths {
		go geth.ConsensusMonitor(errCh, quitCh)
	}

	timeout := time.NewTimer(t)

	defer timeout.Stop()

	select {
	case err := <-errCh:
		Expect(err).Should(BeNil())
	case <-timeout.C:
		for i := 0; i < len(geths); i++ {
			quitCh <- struct{}{}
		}
	}
}

func TestIstanbul(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Istanbul Test Suite")
}
