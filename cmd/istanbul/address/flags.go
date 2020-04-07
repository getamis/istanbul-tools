package address

import (
	"github.com/urfave/cli"
)

var (
	nodeKeyHexFlag = cli.StringFlag{
		Name:  "nodekeyhex",
		Usage: "Node key as hex",
	}
)
