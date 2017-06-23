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
		Usage:     "To decode a Istanbul extraData",
		ArgsUsage: "<extra data>",
		Flags: []cli.Flag{
			ExtraDataFlag,
		},
		Description: `
The extraData command will decode extraData for the given input which should be a hex string.
`,
	}

	encodeCommand = cli.Command{
		Action:    encode,
		Name:      "encode",
		Usage:     "To encode a Istanbul extraData",
		ArgsUsage: "<config file>",
		Flags: []cli.Flag{
			ConfigFlag,
		},
		Description: `
The extraData command will encode Istanbul extraData for the given input file.

Example of input file can refer to example/config.toml.
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
