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
	"github.com/ethereum/go-ethereum/rlp"
	atypes "github.com/getamis/go-ethereum/core/types"
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
		ArgsUsage: "<config file>",
		Flags: []cli.Flag{
			ConfigFlag,
		},
		Description: `
This command encodes vanity and validators to extraData. Please refer to example/config.toml.
`,
	}
)

func encode(ctx *cli.Context) error {
	path := ctx.String(ConfigFlag.Name)
	if len(path) == 0 {
		return cli.NewExitError("Must supply config file", 0)
	}

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

	if len(newVanity) < atypes.IstanbulExtraVanity {
		newVanity = append(newVanity, bytes.Repeat([]byte{0x00}, atypes.IstanbulExtraVanity-len(newVanity))...)
	}
	newVanity = newVanity[:atypes.IstanbulExtraVanity]

	ist := &atypes.IstanbulExtra{
		Validators:    validators,
		Seal:          make([]byte, atypes.IstanbulExtraSeal),
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

	istanbulExtra, err := atypes.ExtractIstanbulExtra(&atypes.Header{Extra: extra})
	if err != nil {
		return err
	}
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
