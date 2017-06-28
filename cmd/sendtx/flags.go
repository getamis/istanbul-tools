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
	"github.com/urfave/cli"
)

var (
	NodeAddrFlag = cli.StringFlag{
		Name:  "addr",
		Usage: "Address of geth node",
	}
	AdminFlag = cli.StringFlag{
		Name:  "admin",
		Usage: "Hex string of admin's private key",
	}
	AccountsFlag = cli.IntFlag{
		Name:  "number",
		Usage: "Number of prepare accounts",
		Value: 4,
	}
	TxsCountFlag = cli.IntFlag{
		Name:  "count",
		Usage: "Number of txs",
		Value: 1000,
	}
	SendPeriodFlag = cli.IntFlag{
		Name:  "period",
		Usage: "Period in seconds to send txs",
		Value: 60,
	}
)
