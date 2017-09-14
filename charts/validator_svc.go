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

type ValidatorServiceChart struct {
	name      string
	chartPath string
	args      []string
}

func NewValidatorServiceChart(name string, args []string) *ValidatorServiceChart {
	chartPath := filepath.Join(chartBasePath, "validator-service")

	chart := &ValidatorServiceChart{
		name:      "validator-svc-" + name,
		args:      args,
		chartPath: chartPath,
	}

	chart.Override("nameOverride", name)
	chart.Override("service.type", "LoadBalancer")
	chart.Override("app", "validator-"+name)

	return chart
}

func (chart *ValidatorServiceChart) Override(key, value string) {
	chart.args = append(chart.args, fmt.Sprintf("%s=%s", key, value))
}

func (chart *ValidatorServiceChart) Install(debug bool) error {
	return installRelease(
		chart.name,
		chart.args,
		chart.chartPath,
		debug,
	)
}

func (chart *ValidatorServiceChart) Uninstall() error {
	return uninstallRelease(chart.name)
}

func (chart *ValidatorServiceChart) Name() string {
	return chart.name
}
