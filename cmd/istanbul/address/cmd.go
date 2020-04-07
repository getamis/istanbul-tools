package address

import (
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli"
)

var (
	AddressCommand = cli.Command{
		Name:   "address",
		Action: address,
		Usage:  "Extract validator address",
		Flags: []cli.Flag{
			nodeKeyHexFlag,
		},
		Description: `Extract validator address`,
	}
)

func address(ctx *cli.Context) error {
	if !ctx.IsSet(nodeKeyHexFlag.Name) {
		return cli.NewExitError("Must supply nodekey hex", 10)
	}
	nodeKeyHex := ctx.String(nodeKeyHexFlag.Name)
	privKey, err := crypto.HexToECDSA(nodeKeyHex)
	if err != nil {
		return err
	}
	fmt.Printf("0x%s\n", hex.EncodeToString(crypto.PubkeyToAddress(privKey.PublicKey).Bytes()))
	return nil
}
