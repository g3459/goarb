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
	// sl := res[0].([]struct {
	// 	Slot0 *big.Int "json:\"slot0\""
	// 	Slot1 *big.Int "json:\"slot1\""
	// 	Slot2 *big.Int "json:\"slot2\""
	// })
	// slp := make([]Pool, len(sl))
	// for i := range sl {
	// 	slp[i] = sl[i]
	// }
	return res[0]
}

func routesDecoder(result interface{}) interface{} {
	res, err := interfaces.RouterABI.Unpack("findRoutes", utils.DecodeHex(*result.(*string)))
	if err != nil {
		return err
	}
	return res[0]
}

func bigIntDecoder(result interface{}) interface{} {
	b := new(big.Int).SetBytes(utils.DecodeHex(*result.(*string)))
	return b
}

func uint64Decoder(result interface{}) interface{} {
	return bigIntDecoder(result).(*big.Int).Uint64()
}
