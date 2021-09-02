package address

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/urfave/cli"
)

var (
	AddressCommand = cli.Command{
		Name:   "address",
		Action: address,
		Usage:  "Extract validator address",
		Flags: []cli.Flag{
			nodeKeyHexFlag,
			nodeIdHexFlag,
		},
		Description: `Extract validator address`,
	}
)

func address(ctx *cli.Context) error {
	if (!ctx.IsSet(nodeKeyHexFlag.Name) && !ctx.IsSet(nodeIdHexFlag.Name)) || (ctx.IsSet(nodeKeyHexFlag.Name) && ctx.IsSet(nodeIdHexFlag.Name)) {
		return cli.NewExitError("Must supply nodekey or nodeid hex", 10)
	}
	var publicKey ecdsa.PublicKey
	if ctx.IsSet(nodeKeyHexFlag.Name) {
		nodeKeyHex := ctx.String(nodeKeyHexFlag.Name)
		privKey, err := crypto.HexToECDSA(nodeKeyHex)
		if err != nil {
			return err
		}
		publicKey = privKey.PublicKey
	}
	if ctx.IsSet(nodeIdHexFlag.Name) {
		pubKey, err := enode.HexPubkey(ctx.String(nodeIdHexFlag.Name))
		if err != nil {
			return err
		}
		publicKey = *pubKey
	}
	fmt.Printf("0x%s\n", hex.EncodeToString(crypto.PubkeyToAddress(publicKey).Bytes()))
	return nil
}
