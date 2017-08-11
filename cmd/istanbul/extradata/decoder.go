package extradata

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

func Decode(extraData string) ([]byte, *types.IstanbulExtra, error) {
	extra, err := hexutil.Decode(extraData)
	if err != nil {
		return nil, nil, err
	}

	istanbulExtra, err := types.ExtractIstanbulExtra(&types.Header{Extra: extra})
	if err != nil {
		return nil, nil, err
	}
	return extra[:types.IstanbulExtraVanity], istanbulExtra, nil
}
