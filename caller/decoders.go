package caller

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

func sliceDecoder(result interface{}) interface{} {
	return *result.(*[]interface{})
}

func stringDecoder(result interface{}) interface{} {
	// fmt.Println("\n\nString:", *result.(*string))
	return *result.(*string)
}

func mapStringDecoder(result interface{}) interface{} {
	return *result.(*map[string]interface{})
}

func allTokensDecoder(result interface{}) interface{} {
	//fmt.Println(*result.(*string))
	dec, _ := hexutil.Decode(*result.(*string))
	res, _ := routerABI.Unpack("allTokensWithBalances", dec)
	if len(res) == 0 {
		return nil
	}
	routesRaw := res[0].([][][]struct {
		AmOut *big.Int "json:\"amOut\""
		Calls []uint8  "json:\"calls\""
	})
	routes := make([][][]Route, len(routesRaw))
	for i := range routes {
		routes[i] = make([][]Route, len(routesRaw[i]))
		for j := range routes[i] {
			routes[i][j] = make([]Route, len(routesRaw[i][j]))
			for k := range routes[i][j] {
				routes[i][j][k] = Route(routesRaw[i][j][k])
			}
		}
	}
	return routes
}

func singleTokenDecoder(result interface{}) interface{} {
	dec, _ := hexutil.Decode(*result.(*string))
	res, _ := routerABI.Unpack("singleToken", dec)
	if len(res) == 0 {
		return nil
	}
	routesRaw := res[0].([]struct {
		AmOut *big.Int "json:\"amOut\""
		//Gas   *big.Int "json:\"gas\""
		Calls []uint8 "json:\"calls\""
	})

	routes := make([]Route, len(routesRaw))
	for k := range routes {
		routes[k] = Route(routesRaw[k])
	}
	return routes
}

func bigIntDecoder(result interface{}) interface{} {
	b, _ := hexutil.DecodeBig(*result.(*string))
	return b
}

func uint64Decoder(result interface{}) interface{} {
	return bigIntDecoder(result).(*big.Int).Uint64()
}
