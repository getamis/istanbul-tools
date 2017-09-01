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

package load

import (
	"testing"

	"github.com/getamis/istanbul-tools/charts"
	"github.com/getamis/istanbul-tools/common"

	"github.com/getamis/istanbul-tools/tests"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TPS-01: Large amount of transactions", func() {
	tests.CaseTable("with number of validators",
		func(numberOfValidators int) {
			tests.CaseTable("with gas limit",
				func(gaslimit int) {
					tests.CaseTable("with txpool size",
						func(txpoolSize int) {
							runTests(numberOfValidators, gaslimit, txpoolSize)
						},

						tests.Case("2048", 2048),
						tests.Case("10240", 10240),
					)
				},

				tests.Case("21000*1000", 21000*1000),
				tests.Case("21000*3000", 21000*3000),
			)
		},

		tests.Case("4 validators", 4),
		tests.Case("7 validators", 7),
		tests.Case("10 validators", 10),
	)
})

func runTests(numberOfValidators int, gaslimit int, txpoolSize int) {
	Describe("", func() {
		var (
			genesisChart     tests.ChartInstaller
			staticNodesChart tests.ChartInstaller
		)

		BeforeEach(func() {
			_, nodekeys, addrs := common.GenerateKeys(numberOfValidators)
			genesisChart = charts.NewGenesisChart(addrs, uint64(gaslimit))
			Expect(genesisChart.Install(false)).To(BeNil())

			staticNodesChart = charts.NewStaticNodesChart(nodekeys, common.GenerateIPs(len(nodekeys)))
			Expect(staticNodesChart.Install(false)).To(BeNil())
		})

		AfterEach(func() {
			Expect(genesisChart.Uninstall()).To(BeNil())
			Expect(staticNodesChart.Uninstall()).To(BeNil())
		})

		It("", func() {
		})
	})
}

func IstanbulLoadTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Istanbul Load Test Suite")
}
