package caller

import (
	"math/big"

	"github.com/g3459/goarb/contracts/interfaces"
	"github.com/g3459/goarb/utils"
)

func poolsDecoder(result interface{}) interface{} {
	res, err := interfaces.PoolFinderABI.Unpack("findPools", utils.DecodeHex(*result.(*string)))
	if err != nil {
		return err
	}
	return res[0]
}

func routesDecoder(result interface{}) interface{} {
	res, err := interfaces.RouterABI.Unpack("findRoutes", utils.DecodeHex(*result.(*string)))
	if err != nil {
		return err
	}
	amounts := res[0].([]*big.Int)
	calls := res[1].([][]byte)
	routes := make([]Route, len(amounts))
	for i := range routes {
		routes[i].AmOut = amounts[i]
		routes[i].Calls = calls[i]
	}
	return routes
}

func bigIntDecoder(result interface{}) interface{} {
	b := new(big.Int).SetBytes(utils.DecodeHex(*result.(*string)))
	return b
}

func uint64Decoder(result interface{}) interface{} {
	return bigIntDecoder(result).(*big.Int).Uint64()
}
