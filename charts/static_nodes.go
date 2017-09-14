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

package charts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/getamis/istanbul-tools/common"
)

type StaticNodesChart struct {
	name            string
	chartPath       string
	staticNodesFile string
	args            []string
}

func NewStaticNodesChart(nodekeys []string, ipAddrs []string) *StaticNodesChart {
	chartPath := filepath.Join(chartBasePath, "static-nodes")
	staticNodesPath := filepath.Join(chartPath, ".static-nodes")
	err := os.MkdirAll(staticNodesPath, 0700)
	if err != nil {
		log.Error("Failed to create dir", "dir", staticNodesPath, "err", err)
	}

	if len(nodekeys) != len(ipAddrs) {
		log.Error("The number of nodekeys and the number of IP address should be equal", "nodekeys", len(nodekeys), "ips", len(ipAddrs))
		return nil
	}

	chart := &StaticNodesChart{
		name:            "static-nodes",
		chartPath:       chartPath,
		staticNodesFile: common.GenerateStaticNodesAt(staticNodesPath, nodekeys, ipAddrs),
	}

	relPath := strings.Replace(chart.staticNodesFile, chartPath+"/", "", 1)
	chart.Override("fileName", relPath)

	return chart
}

func (chart *StaticNodesChart) Override(key, value string) {
	chart.args = append(chart.args, fmt.Sprintf("%s=%s", key, value))
}

func (chart *StaticNodesChart) Install(debug bool) error {
	defer os.RemoveAll(filepath.Dir(chart.staticNodesFile))

	return installRelease(
		chart.name,
		chart.args,
		chart.chartPath,
		debug,
	)
}

func (chart *StaticNodesChart) Uninstall() error {
	return uninstallRelease(chart.name)
}
