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

package extra

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/naoina/toml"
	"github.com/urfave/cli"
)

var (
	ExtraCommand = cli.Command{
		Name:  "extra",
		Usage: "qbft extraData manipulation",
		Subcommands: []cli.Command{
			cli.Command{
				Action:    decode,
				Name:      "decode",
				Usage:     "To decode an qbft extraData",
				ArgsUsage: "<extra data>",
				Flags: []cli.Flag{
					extraDataFlag,
				},
				Description: `
		This command decodes extraData to vanity and validators.
		`,
			},
			cli.Command{
				Action:    encode,
				Name:      "encode",
				Usage:     "To encode an qbft extraData",
				ArgsUsage: "<config file> or \"0xValidator1,0xValidator2...\"",
				Flags: []cli.Flag{
					configFlag,
					validatorsFlag,
					vanityFlag,
				},
				Description: `
		This command encodes vanity and validators to extraData. Please refer to example/config.toml.
		`,
			},
		},
	}
)

func encode(ctx *cli.Context) error {
	path := ctx.String(configFlag.Name)
	validators := ctx.String(validatorsFlag.Name)
	if len(path) == 0 && len(validators) == 0 {
		return cli.NewExitError("Must supply config file or enter validators", 0)
	}

	if len(path) != 0 {
		extraData, err := fromConfig(path)
		if err != nil {
			return cli.NewExitError("Failed to encode from config data", 0)
		}
		fmt.Println("Encoded qbft extra-data:", extraData)
	}

	if len(validators) != 0 {
		extraData, err := fromRawData(ctx.String(vanityFlag.Name), validators)
		if err != nil {
			return cli.NewExitError("Failed to encode from flags", 0)
		}
		fmt.Println("Encoded qbft extra-data:", extraData)
	}
	return nil
}

func fromRawData(vanity string, validators string) (string, error) {
	vs := splitAndTrim(validators)

	addrs := make([]common.Address, len(vs))
	for i, v := range vs {
		addrs[i] = common.HexToAddress(v)
	}
	return Encode(vanity, addrs)
}

func fromConfig(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", cli.NewExitError(fmt.Sprintf("Failed to read config file: %v", err), 1)
	}
	defer file.Close()

	var config struct {
		Vanity     string
		Validators []common.Address
	}

	if err := toml.NewDecoder(file).Decode(&config); err != nil {
		return "", cli.NewExitError(fmt.Sprintf("Failed to parse config file: %v", err), 2)
	}

	return Encode(config.Vanity, config.Validators)
}

func decode(ctx *cli.Context) error {
	if !ctx.IsSet(extraDataFlag.Name) {
		return cli.NewExitError("Must supply extra data", 10)
	}

	extraString := ctx.String(extraDataFlag.Name)
	qbftExtra, err := Decode(extraString)
	if err != nil {
		return err
	}

	fmt.Println("vanity: ", "0x"+common.Bytes2Hex(qbftExtra.VanityData))

	for _, v := range qbftExtra.Validators {
		fmt.Println("validator: ", v.Hex())
	}

	for _, seal := range qbftExtra.CommittedSeal {
		fmt.Println("committed seal: ", "0x"+common.Bytes2Hex(seal))
	}

	if qbftExtra.Vote != nil {
		fmt.Printf("validator vote: address=0x%x, type=%x\n", qbftExtra.Vote.RecipientAddress, qbftExtra.Vote.VoteType)
	}

	fmt.Println("round: ", qbftExtra.Round)

	return nil
}
