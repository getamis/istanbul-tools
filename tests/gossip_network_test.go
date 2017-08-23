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

var _ = Describe("TFS-07: Gossip Network", func() {
	const (
		numberOfValidators = 4
	)
	var (
		blockchain container.Blockchain
	)

	BeforeEach(func() {
		blockchain = container.NewDefaultBlockchain(numberOfValidators)
		Expect(blockchain.Start(false)).To(BeNil())
	})

	AfterEach(func() {
		Expect(blockchain.Stop(false)).To(BeNil())
		blockchain.Finalize()
	})

	It("TFS-07-01: Gossip Network", func(done Done) {
		By("Check peer count", func() {
			for _, geth := range blockchain.Validators() {
				c := geth.NewIstanbulClient()
				peers, e := c.AdminPeers(context.Background())
				Expect(e).To(BeNil())
				Ω(len(peers)).Should(BeNumerically("<=", 2))
			}
		})

		By("Checking blockchain progress", func() {
			v0 := blockchain.Validators()[0]
			c0 := v0.NewClient()
			ticker := time.NewTicker(time.Millisecond * 100)
			for _ = range ticker.C {
				b, e := c0.BlockByNumber(context.Background(), nil)
				Expect(e).To(BeNil())
				// Check if new blocks are getting generated
				if b.Number().Int64() > 1 {
					ticker.Stop()
					break
				}
			}
		})
		close(done)
	}, 240)
})
