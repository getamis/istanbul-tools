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
	"math/big"
	"sync"

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
		blockchain = container.NewDefaultBlockchain(numberOfValidators)
		Expect(blockchain.Start(true)).To(BeNil())
	})

	AfterEach(func() {
		blockchain.Stop(true) // This will return container not found error since we stop one
		blockchain.Finalize()
	})

	It("TFS-01-03: Peer connection", func(done Done) {
		expectedPeerCount := len(blockchain.Validators()) - 1
		waitFor(blockchain.Validators(), func(v container.Ethereum, wg *sync.WaitGroup) {
			Expect(v.WaitForPeersConnected(expectedPeerCount)).To(BeNil())
			wg.Done()
		})

		close(done)
	}, 20)

	It("TFS-01-04: Consensus progress", func(done Done) {
		const (
			targetBlockHeight = 10
			maxBlockPeriod    = 3
		)

		By("Wait for consensus progress", func() {
			waitFor(blockchain.Validators(), func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.WaitForBlockHeight(targetBlockHeight)).To(BeNil())
				wg.Done()
			})
		})

		By("Check the block period should less than 3 seconds", func() {
			errc := make(chan error, len(blockchain.Validators()))
			for _, geth := range blockchain.Validators() {
				go func(geth container.Ethereum) {
					c := geth.NewClient()
					lastBlockTime := int64(0)
					for i := 1; i <= targetBlockHeight; i++ {
						header, err := c.HeaderByNumber(context.Background(), big.NewInt(int64(i)))
						if err != nil {
							errc <- err
							return
						}
						if lastBlockTime != 0 {
							diff := header.Time.Int64() - lastBlockTime
							if diff > maxBlockPeriod {
								errc <- errors.New("Invalid block period.")
								return
							}
						}
						lastBlockTime = header.Time.Int64()
					}
					errc <- nil
				}(geth)
			}

			for i := 0; i < len(blockchain.Validators()); i++ {
				err := <-errc
				Expect(err).To(BeNil())
			}
		})
		close(done)
	}, 60)
})
