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
})
