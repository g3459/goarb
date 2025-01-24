package caller

import (
	"log"
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
	log.Println((*result.(*interface{})).(string))

	res, err := RouterABI.Unpack("findRoutes", (utils.DecodeHex((*result.(*interface{})).(string))))
	if err != nil {
		return err
	}
	amounts := res[0].([]*big.Int)
	calls := res[1].([][]byte)
	gasUsage := res[2].([]uint64)
	routes := make([]Route, len(amounts))
	for i := range routes {
		routes[i].AmOut = amounts[i]
		routes[i].Calls = calls[i]
		routes[i].GasUsage = gasUsage[i]
	}
	return routes
}

func bigIntDecoder(result interface{}) interface{} {
	b := new(big.Int).SetBytes((utils.DecodeHex((*result.(*interface{})).(string))))
	return b
}

func uint64Decoder(result interface{}) interface{} {
	return bigIntDecoder(result).(*big.Int).Uint64()
}
