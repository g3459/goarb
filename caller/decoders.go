package caller

import (
	"math/big"

	"github.com/g3459/goarb/utils"
)

func findPoolsDecoder(result interface{}) interface{} {
	res, err := PoolFinderABI.Unpack("findPools", (utils.DecodeHex((*result.(*interface{})).(string))))
	if err != nil {
		return err
	}
	return res[0]
}

func findPoolsCheckBlockNumberDecoder(result interface{}) interface{} {
	res, err := PoolFinderABI.Unpack("findPoolsCheckBlockNumber", (utils.DecodeHex((*result.(*interface{})).(string))))
	if err != nil {
		return err
	}
	return res
}

func findRoutesDecoder(result interface{}) interface{} {
	res, err := RouterABI.Unpack("findRoutes", (utils.DecodeHex((*result.(*interface{})).(string))))
	if err != nil {
		return err
	}
	calls := res[0].([][]byte)
	return calls
}

func bigIntDecoder(result interface{}) interface{} {
	b := new(big.Int).SetBytes((utils.DecodeHex((*result.(*interface{})).(string))))
	return b
}

func uint64Decoder(result interface{}) interface{} {
	return bigIntDecoder(result).(*big.Int).Uint64()
}
