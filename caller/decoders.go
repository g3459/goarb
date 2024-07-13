package caller

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/g3459/goarb/contracts/interfaces"
)

func sliceDecoder(result interface{}) interface{} {
	return *result.(*[]interface{})
}

func stringDecoder(result interface{}) interface{} {
	return *result.(*string)
}

func mapStringDecoder(result interface{}) interface{} {
	return *result.(*map[string]interface{})
}

func poolsDecoder(result interface{}) interface{} {
	dec, _ := hexutil.Decode(*result.(*string))
	res, _ := interfaces.PoolFinderABI.Unpack("findPools", dec)
	return res[0]
}

func routesDecoder(result interface{}) interface{} {
	//fmt.Println(*result.(*string))

	dec, _ := hexutil.Decode(*result.(*string))
	res, _ := interfaces.RouterABI.Unpack("findRoutes", dec)
	// if len(res) == 0 {
	// 	return nil
	// }
	// routesRaw := res[0].([][][]struct {
	// 	AmOut *big.Int "json:\"amOut\""
	// 	Calls []uint8  "json:\"calls\""
	// })
	// routes := make([][][]Route, len(routesRaw))
	// for i := range routes {
	// 	routes[i] = make([][]Route, len(routesRaw[i]))
	// 	for j := range routes[i] {
	// 		routes[i][j] = make([]Route, len(routesRaw[i][j]))
	// 		for k := range routes[i][j] {
	// 			routes[i][j][k] = Route(routesRaw[i][j][k])
	// 		}
	// 	}
	// }
	return res[0]
}

func bigIntDecoder(result interface{}) interface{} {
	b := new(big.Int).SetBytes(hexutil.MustDecode(*result.(*string)))
	return b
}

func uint64Decoder(result interface{}) interface{} {
	return bigIntDecoder(result).(*big.Int).Uint64()
}
