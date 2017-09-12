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

package k8s

import (
	"fmt"

	"github.com/getamis/istanbul-tools/charts"
	"github.com/getamis/istanbul-tools/common"
	"github.com/getamis/istanbul-tools/genesis"
)

func ExampleK8SEthereum() {
	_, nodekeys, addrs := common.GenerateKeys(1)
	genesisChart := charts.NewGenesisChart(addrs, genesis.InitGasLimit)
	if err := genesisChart.Install(false); err != nil {
		fmt.Println(err)
		return
	}
	defer genesisChart.Uninstall()

	staticNodesChart := charts.NewStaticNodesChart(nodekeys, common.GenerateIPs(len(nodekeys)))
	if err := staticNodesChart.Install(false); err != nil {
		fmt.Println(err)
		return
	}
	defer staticNodesChart.Uninstall()

	geth := NewEthereum(
		ImageRepository("quay.io/amis/geth"),
		ImageTag("istanbul_develop"),

		Name("test"),
		ServiceType("LoadBalancer"),
		IPAddress("10.0.1.100"),
		NodeKeyHex(common.RandomHex()[2:]),
	)

	err := geth.Start()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = geth.Stop()
	if err != nil {
		fmt.Println(err)
		return
	}
}
