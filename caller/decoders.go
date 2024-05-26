package caller

import (
	"math/big"

	"github.com/g3459/goarb/utils"
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

func allTokensDecoder(result interface{}) interface{} {
	res, _ := routerABI.Unpack("allTokensWithBalances", utils.HexToBytes(*result.(*string)))
	if len(res) == 0 {
		return nil
	}
	routesRaw := res[0].([][]struct {
		AmIn   *big.Int "json:\"amIn\""
		Routes []struct {
			AmOut *big.Int "json:\"amOut\""
			//Gas   *big.Int "json:\"gas\""
			Calls []uint8 "json:\"calls\""
		} "json:\"routes\""
	})
	routes := make([][]Routes, len(routesRaw))
	for i := range routes {
		routes[i] = make([]Routes, len(routesRaw[i]))
		for j := range routes[i] {
			routes[i][j].AmIn = routesRaw[i][j].AmIn
			routes[i][j].Routes = make([]Route, len(routesRaw[i][j].Routes))
			for k := range routes[i][j].Routes {
				routes[i][j].Routes[k] = Route(routesRaw[i][j].Routes[k])
			}
		}
	}
	return routes
}

func singleTokenDecoder(result interface{}) interface{} {
	res, _ := routerABI.Unpack("singleToken", utils.HexToBytes(*result.(*string)))
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
	return new(big.Int).SetBytes(utils.HexToBytes(*result.(*string)))
}

func uint64Decoder(result interface{}) interface{} {
	return bigIntDecoder(result).(*big.Int).Uint64()
}
