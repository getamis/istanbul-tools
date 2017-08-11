package extradata

import (
	"bytes"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rlp"
	atypes "github.com/getamis/go-ethereum/core/types"
)

func Encode(vanity string, validators []common.Address) (string, error) {
	newVanity, err := hexutil.Decode(vanity)
	if err != nil {
		return "", err
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
		return "", err
	}

	return "0x" + common.Bytes2Hex(append(newVanity, payload...)), nil
}
