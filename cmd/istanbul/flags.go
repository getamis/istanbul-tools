package main

import (
	"github.com/urfave/cli"
)

var (
	ConfigFlag = cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}

	ExtraDataFlag = cli.StringFlag{
		Name:  "extradata",
		Usage: "Hex string for RLP encoded Istanbul extraData",
	}
)
