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
	"bytes"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/naoina/toml"
	"github.com/urfave/cli"
)

var (
	decodeCommand = cli.Command{
		Action:    decode,
		Name:      "decode",
		Usage:     "To decode an Istanbul extraData",
		ArgsUsage: "<extra data>",
		Flags: []cli.Flag{
			ExtraDataFlag,
		},
		Description: `
This command decodes extraData to vanity and validators.
`,
	}

	encodeCommand = cli.Command{
		Action:    encode,
		Name:      "encode",
		Usage:     "To encode an Istanbul extraData",
		ArgsUsage: "<config file> or \"0xValidator1,0xValidator2...\"",
		Flags: []cli.Flag{
			ConfigFlag,
			ValidatorsFlag,
			VanityFlag,
		},
		Description: `
This command encodes vanity and validators to extraData. Please refer to example/config.toml.
`,
	}
)

func encode(ctx *cli.Context) error {
	path := ctx.String(ConfigFlag.Name)
	validators := ctx.String(ValidatorsFlag.Name)
	if len(path) == 0 && len(validators) == 0 {
		return cli.NewExitError("Must supply config file or enter validators", 0)
	}

	if len(path) != 0 {
		if err := fromConfig(path); err != nil {
			return cli.NewExitError("Failed to encode from config data", 0)
		}
	}

	if len(validators) != 0 {
		if err := fromRawData(ctx.String(VanityFlag.Name), validators); err != nil {
			return cli.NewExitError("Failed to encode from flags", 0)
		}
	}
	return nil
}

func fromRawData(vanity string, validators string) error {
	vs := splitAndTrim(validators)

	addrs := make([]common.Address, len(vs))
	for i, v := range vs {
		addrs[i] = common.HexToAddress(v)
	}
	return encodeExtraData(vanity, addrs)
}

func fromConfig(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to read config file: %v", err), 1)
	}
	defer file.Close()

	var config struct {
		Vanity     string
		Validators []common.Address
	}

	if err := toml.NewDecoder(file).Decode(&config); err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to parse config file: %v", err), 2)
	}

	return encodeExtraData(config.Vanity, config.Validators)
}

func decode(ctx *cli.Context) error {
	if !ctx.IsSet(ExtraDataFlag.Name) {
		return cli.NewExitError("Must supply extra data", 10)
	}

	return decodeExtraData(ctx.String(ExtraDataFlag.Name))
}

func encodeExtraData(vanity string, validators []common.Address) error {
	newVanity, err := hexutil.Decode(vanity)
	if err != nil {
		return err
	}

	if len(newVanity) < types.IstanbulExtraVanity {
		newVanity = append(newVanity, bytes.Repeat([]byte{0x00}, types.IstanbulExtraVanity-len(newVanity))...)
	}
	newVanity = newVanity[:types.IstanbulExtraVanity]

	ist := &types.IstanbulExtra{
		Validators:    validators,
		Seal:          make([]byte, types.IstanbulExtraSeal),
		CommittedSeal: [][]byte{},
	}

	payload, err := rlp.EncodeToBytes(&ist)
	if err != nil {
		return err
	}
	fmt.Println("Encoded Istanbul extra-data:", "0x"+common.Bytes2Hex(append(newVanity, payload...)))

	return nil
}

func decodeExtraData(extraData string) error {
	extra, err := hexutil.Decode(extraData)
	if err != nil {
		return err
	}

	istanbulExtra, err := types.ExtractIstanbulExtra(&types.Header{Extra: extra})
	if err != nil {
		return err
	}

	fmt.Println("vanity: ", "0x"+common.Bytes2Hex(extra[:types.IstanbulExtraVanity]))

	for _, v := range istanbulExtra.Validators {
		fmt.Println("validator: ", v.Hex())
	}

	if len(istanbulExtra.Seal) != 0 {
		fmt.Println("seal:", "0x"+common.Bytes2Hex(istanbulExtra.Seal))
	}

	for _, seal := range istanbulExtra.CommittedSeal {
		fmt.Println("committed seal: ", "0x"+common.Bytes2Hex(seal))
	}

	return nil
}
