package utils

import (
	"math/big"
	"slices"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func SignTx(txData *types.DynamicFeeTx, privateKey *common.Hash) string {
	ecdsapk, _ := crypto.ToECDSA((*privateKey)[:])
	tx, _ := types.SignNewTx(ecdsapk, types.NewCancunSigner(txData.ChainID), txData)
	data, _ := tx.MarshalBinary()
	return hexutil.Encode(data)
}

func RouteGas(calls []byte) uint64 {
	gas := uint64(30000)
	for i := 0; i < len(calls); i += 32 {
		if calls[i+4] == 2 {
			gas += 285000
		} else {
			gas += 100000
		}
	}
	return gas
}

func AccessListForCalls(calls []byte) types.AccessList {
	al := make([]types.AccessTuple, len(calls)/32)
	addrs := make([]common.Address, len(al))
	n := 0
	for i := 0; i < len(al); i++ {
		byteIx := i * 32
		addr := common.Address(calls[byteIx+12 : byteIx+32])
		if !slices.Contains(addrs, addr) {
			al[n].Address = addr
			addrs[n] = addr
			if calls[(i*32)+4] == 1 {
				al[n].StorageKeys = []common.Hash{common.BigToHash(big.NewInt(3))}
			} else {
				al[n].StorageKeys = []common.Hash{common.BigToHash(big.NewInt(0))}
			}
			n++
		}
	}
	return al[:n]
}
