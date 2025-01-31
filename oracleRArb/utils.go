package main

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func AccessListForCalls(calls []byte) types.AccessList {
	al := []types.AccessTuple{}
	for i := 0; i <= len(calls)-0x20; i += 0x20 {
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
		var slot uint8
		pId := calls[i+4]
		if pId == 0 {
			slot = 0
		} else if pId == 1 {
			slot = 8
		} else if pId == 2 {
			slot = 3
		}
		h := common.Hash{}
		h[31] = slot
		al = append(al, types.AccessTuple{Address: addr, StorageKeys: []common.Hash{h}})
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

type any = interface{}

var logLevel int

var logmu sync.Mutex

func Log(level int, params ...any) {
	logmu.Lock()
	if logLevel >= level || logLevel == 0 {
		if level < 0 {
			s := fmt.Sprintln(params...)
			panic(s)
		} else {
			fmt.Println(params...)
		}
	}
	logmu.Unlock()
}
