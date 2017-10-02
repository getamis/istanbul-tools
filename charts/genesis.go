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
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/getamis/istanbul-tools/genesis"
)

type GenesisChart struct {
	name        string
	chartPath   string
	genesisFile string
	args        []string
}

func NewGenesisChart(validators []common.Address, allocs []common.Address, gasLimit uint64) *GenesisChart {
	chartPath := filepath.Join(chartBasePath, "genesis-block")
	genesisPath := filepath.Join(chartPath, ".genesis")
	err := os.MkdirAll(genesisPath, 0700)
	if err != nil {
		log.Error("Failed to create dir", "dir", genesisPath, "err", err)
		return nil
	}

	chart := &GenesisChart{
		name:      "genesis-block",
		chartPath: chartPath,
		genesisFile: genesis.NewFileAt(
			genesisPath,
			false,
			genesis.Validators(validators...),
			genesis.GasLimit(gasLimit),
			genesis.Alloc(append(validators, allocs...), new(big.Int).Exp(big.NewInt(10), big.NewInt(50), nil)),
		),
	}

	relPath := strings.Replace(chart.genesisFile, chartPath+"/", "", 1)
	chart.Override("genesisFileName", relPath)

	return chart
}

func (chart *GenesisChart) Override(key, value string) {
	chart.args = append(chart.args, fmt.Sprintf("%s=%s", key, value))
}

func (chart *GenesisChart) Install(debug bool) error {
	defer os.RemoveAll(filepath.Dir(chart.genesisFile))

	return installRelease(
		chart.name,
		chart.args,
		chart.chartPath,
		debug,
	)
}

func (chart *GenesisChart) Uninstall() error {
	return uninstallRelease(chart.name)
}
