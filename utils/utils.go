package utils

import (
	"encoding/hex"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func DecodeHex(s string) []byte {
	if len(s) >= 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
		s = s[2:]
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func SignTx(txData *types.DynamicFeeTx, privateKey *common.Hash) string {
	tx, _ := types.SignNewTx(crypto.ToECDSAUnsafe((*privateKey)[:]), types.NewCancunSigner(txData.ChainID), txData)
	data, _ := tx.MarshalBinary()
	return hexutil.Encode(data)
}
