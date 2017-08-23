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

var _ = Describe("Dynamic validators addition/removal testing", func() {
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
		Expect(blockchain.Stop(false)).To(BeNil())
		blockchain.Finalize()
	})

	It("TFS-02-01 Add validators", func() {
		testValidators := 3

		By("Ensure that numbers of validator is equal than $numberOfValidators", func() {
			for _, v := range blockchain.Validators() {
				client := v.NewIstanbulClient()
				validators, err := client.GetValidators(context.Background(), nil)
				Expect(err).Should(BeNil())
				Expect(len(validators)).Should(BeNumerically("==", numberOfValidators))
			}
		})

		_, err := blockchain.AddValidators(testValidators)
		Expect(err).Should(BeNil())

		By("Ensure that consensus is working in 50 seconds", func() {
			Expect(blockchain.EnsureConsensusWorking(blockchain.Validators(), 50*time.Second)).Should(BeNil())
		})
		for _, v := range blockchain.Validators() {
			client := v.NewIstanbulClient()
			validators, err := client.GetValidators(context.Background(), nil)
			Expect(err).Should(BeNil())
			Expect(len(validators)).Should(BeNumerically("==", numberOfValidators+testValidators))
		}
	})
})
