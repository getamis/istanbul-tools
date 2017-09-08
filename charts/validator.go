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
	"path/filepath"
)

type ValidatorChart struct {
	name      string
	chartPath string
	args      []string
}

func NewValidatorChart(name string, args []string) *ValidatorChart {
	chartPath := filepath.Join(chartBasePath, "validator")

	chart := &ValidatorChart{
		name:      "validator-" + name,
		args:      args,
		chartPath: chartPath,
	}

	return chart
}

func (chart *ValidatorChart) Override(key, value string) {
	chart.args = append(chart.args, fmt.Sprintf("%s=%s", key, value))
}

func (chart *ValidatorChart) Install(debug bool) error {
	return installRelease(
		chart.name,
		chart.args,
		chart.chartPath,
		debug,
	)
}

func (chart *ValidatorChart) Uninstall() error {
	return uninstallRelease(chart.name)
}

func (chart *ValidatorChart) Name() string {
	return chart.name
}
