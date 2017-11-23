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

package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"

	"github.com/getamis/istanbul-tools/cmd/istanbul/extra"
	"github.com/getamis/istanbul-tools/cmd/istanbul/setup"
	"github.com/getamis/istanbul-tools/cmd/utils"
)

func main() {
	app := utils.NewApp()
	app.Usage = "the istanbul-tools command line interface"

	app.Version = "v1.0.0"
	app.Copyright = "Copyright 2017 The AMIS Authors"
	app.Commands = []cli.Command{
		extra.ExtraCommand,
		setup.SetupCommand,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
