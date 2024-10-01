package utils

import (
	"encoding/hex"
	"math/big"

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
	ecdsapk, _ := crypto.ToECDSA((*privateKey)[:])
	tx, _ := types.SignNewTx(ecdsapk, types.NewCancunSigner(txData.ChainID), txData)
	data, _ := tx.MarshalBinary()
	return hexutil.Encode(data)
}

func ExecuteCallsGas(calls []byte) uint64 {
	return CallsGas(calls) + 50000
}

func CallsGas(calls []byte) uint64 {
	gas := uint64(0)
	for i := 0; i < len(calls); i += 32 {
		if calls[i+4] == 2 {
			gas += 300000
		} else {
			gas += 100000
		}
	}
	return gas
}

func AccessListForCalls(calls []byte) types.AccessList {
	al := []types.AccessTuple{}
	for i := 0; i < len(calls); i += 0x20 {
		addr := common.Address(calls[i+12 : i+32])
		cont := false
		for _, v := range al {
			if v.Address.Cmp(addr) == 0 {
				cont = true
				break
			}
		}
		if cont {
			continue
		}
		var slot int64
		if calls[i+4] != 0 {
			slot = 8
		}
		al = append(al, types.AccessTuple{Address: addr, StorageKeys: []common.Hash{common.BigToHash(big.NewInt(slot))}})
	}
	return al
}

func ToX64Int(n float64) *big.Int {
	fl := big.NewFloat(n)
	i, _ := fl.Mul(fl, new(big.Float).SetInt(new(big.Int).Lsh(big.NewInt(1), 64))).Int(nil)
	return i
}

func PoolDif(calls1 []byte, calls2 []byte) bool {
	for i := 0; i < len(calls1); i += 32 {
		for ii := 0; ii < len(calls2); ii += 32 {
			dif := false
			for j := 0; j < 20; j++ {
				if calls1[i+12+j] != calls2[ii+12+j] {
					dif = true
					break
				}
			}
			if !dif {
				return false
			}
		}
	}
	return true
}
