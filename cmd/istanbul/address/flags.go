package address

import (
	"github.com/urfave/cli"
)

var (
	nodeKeyHexFlag = cli.StringFlag{
		Name:  "nodekeyhex",
		Usage: "Node key as hex",
	}
	nodeIdHexFlag = cli.StringFlag{
		Name:  "nodeidhex",
		Usage: "Public key as hex, usually used as enode address",
	}
)
