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

package reinit

import (
  "strings"

  "github.com/urfave/cli"
)

var (
  nodeKeyFlag = cli.StringFlag{
    Name:  "nodekey",
    Usage: "String of comma separated nodekey values",
  }

  quorumFlag = cli.BoolFlag{
    Name:  "quorum",
    Usage: "Use Quorum",
  }
)

func splitAndTrim(input string) []string {
  result := strings.Split(input, ",")
  for i, r := range result {
    result[i] = strings.TrimSpace(r)
  }
  return result
}
