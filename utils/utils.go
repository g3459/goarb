package utils

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func SignTx(tx *types.Transaction, chainId uint, privateKey string) string {
	ecdsapk, _ := crypto.HexToECDSA(privateKey)
	tx, _ = types.SignTx(tx, types.NewEIP155Signer(big.NewInt(int64(chainId))), ecdsapk)
	buf := new(bytes.Buffer)
	tx.EncodeRLP(buf)
	return hexutil.Encode(buf.Bytes())
}

func RouteGas(calls []byte) (gas uint) {
	gas = 21000
	gas += uint((len(calls) / 24) * 95000)
	// for i := range calls {
	// 	stateSelector := utils.BytesToHex(calls[i].StateSelector[:])
	// 	if stateSelector == "0x3850c7bd" || stateSelector == "0x0902f1ac" {
	// 		gas += 90000
	// 	} else {
	// 		gas += 170000
	// 	}
	// }
	return gas
}
