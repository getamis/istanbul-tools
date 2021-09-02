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
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Consensys/istanbul-tools/container"
	"github.com/Consensys/istanbul-tools/tests"
)

var _ = Describe("TFS-07: Gossip Network", func() {
	const (
		numberOfValidators = 4
	)
	var (
		blockchain container.Blockchain
	)

	BeforeEach(func() {
		blockchain = container.NewDefaultBlockchain(dockerNetwork, numberOfValidators)
		Expect(blockchain.Start(false)).To(BeNil())
	})

	AfterEach(func() {
		Expect(blockchain.Stop(false)).To(BeNil())
		blockchain.Finalize()
	})

	It("TFS-07-01: Gossip Network", func(done Done) {
		By("Check peer count", func() {
			for _, geth := range blockchain.Validators() {
				c := geth.NewClient()
				peers, e := c.AdminPeers(context.Background())
				Expect(e).To(BeNil())
				Ω(len(peers)).Should(BeNumerically("<=", 2))
			}
		})

		By("Checking blockchain progress", func() {
			tests.WaitFor(blockchain.Validators(), func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.WaitForBlocks(3)).To(BeNil())
				wg.Done()
			})
		})

		close(done)
	}, 240)
})
