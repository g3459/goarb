package utils

import (
	"bytes"
	"encoding/hex"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func HexToBytes(val string) []byte {
	if len(val) >= 2 && val[:2] == "0x" {
		val = val[2:]
	}
	if int(len(val)/2)*2 != len(val) {
		val = "0" + val
	}
	res, _ := hex.DecodeString(val)
	return res
}

func HexNumToBytes(val string) []byte {
	return common.TrimLeftZeroes(HexToBytes(val))
}

func BytesToHex(val []byte) string {
	res := hex.EncodeToString(val)
	return "0x" + res
}

func BytesToHexNum(val []byte) string {
	return BytesToHex(common.TrimLeftZeroes(val))
}

func SignTx(tx *types.Transaction, chainId uint, privateKey string) string {
	ecdsapk, _ := crypto.HexToECDSA(privateKey)
	tx, _ = types.SignTx(tx, types.NewEIP155Signer(big.NewInt(int64(chainId))), ecdsapk)
	buf := new(bytes.Buffer)
	tx.EncodeRLP(buf)
	return BytesToHex(buf.Bytes())
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
