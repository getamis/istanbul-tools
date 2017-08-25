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
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/getamis/istanbul-tools/container"
)

var _ = Describe("TFS-05: Byzantine Faulty", func() {

	Context("TFS-05-01: F faulty validators", func() {
		const (
			numberOfNormal = 3
			numberOfFaulty = 1
		)
		var (
			blockchain container.Blockchain
		)
		BeforeEach(func() {
			blockchain = container.NewDefaultBlockchainWithFaulty(numberOfNormal, numberOfFaulty)
			Expect(blockchain.Start(true)).To(BeNil())
		})

		AfterEach(func() {
			Expect(blockchain.Stop(false)).To(BeNil())
			blockchain.Finalize()
		})

		It("Should generate blocks", func(done Done) {

			By("Wait for p2p connection", func() {
				waitFor(blockchain.Validators(), func(geth container.Ethereum, wg *sync.WaitGroup) {
					Expect(geth.WaitForPeersConnected(numberOfNormal + numberOfFaulty - 1)).To(BeNil())
					wg.Done()
				})
			})

			By("Wait for blocks", func() {
				const targetBlockHeight = 3
				waitFor(blockchain.Validators()[:1], func(geth container.Ethereum, wg *sync.WaitGroup) {
					Expect(geth.WaitForBlocks(targetBlockHeight)).To(BeNil())
					wg.Done()
				})
			})

			close(done)
		}, 60)
	})

	Context("TFS-05-01: F+1 faulty validators", func() {
		const (
			numberOfNormal = 2
			numberOfFaulty = 2
		)
		var (
			blockchain container.Blockchain
		)
		BeforeEach(func() {
			blockchain = container.NewDefaultBlockchainWithFaulty(numberOfNormal, numberOfFaulty)
			Expect(blockchain.Start(true)).To(BeNil())
		})

		AfterEach(func() {
			Expect(blockchain.Stop(false)).To(BeNil())
			blockchain.Finalize()
		})

		It("Should not generate blocks", func(done Done) {
			By("Wait for p2p connection", func() {
				waitFor(blockchain.Validators(), func(geth container.Ethereum, wg *sync.WaitGroup) {
					Expect(geth.WaitForPeersConnected(numberOfNormal + numberOfFaulty - 1)).To(BeNil())
					wg.Done()
				})
			})

			By("Wait for blocks", func() {
				// Only check normal validators
				waitFor(blockchain.Validators()[:2], func(geth container.Ethereum, wg *sync.WaitGroup) {
					Expect(geth.WaitForNoBlocks(0, time.Second*30)).To(BeNil())
					wg.Done()
				})
			})
			close(done)
		}, 60)
	})

})
