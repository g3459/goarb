package main

import (
	//"context"

	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/g3459/goarb/caller"
	"github.com/g3459/goarb/utils"
)

type Configuration struct {
	Router       common.Address     `json:"router"`
	Caller       common.Address     `json:"caller"`
	Tokens       []caller.TokenInfo `json:"tokens"`
	EthPricesX64 []*big.Int         `json:"ethPricesX64"`
	WsRpcs       []string           `json:"wsRpcs"`
}

var tokenDecimals = map[common.Address]uint{
	common.HexToAddress("0xc2132d05d31c914a87c6611c10748aeb04b58e8f"): 6,
	common.HexToAddress("0x2791bca1f2de4661ed88a30c99a7a9449aa84174"): 6,
	common.HexToAddress("0x0d500b1d8e8ef31e21c99d1db9a6444d3adf1270"): 18,
	common.HexToAddress("0x7ceb23fd6bc0add59e62ac25578270cff1b9f619"): 18,
	common.HexToAddress("0x8f3cf7ad23cd3cadbd9735aff958023239c6a063"): 18,
	common.HexToAddress("0x1bfd67037b42cf73acf2047067bd4f2c47d9bfd6"): 8,
	common.HexToAddress("0x53e0bca35ec356bd5dddfebbd1fc0fd03fabad39"): 18,
	common.HexToAddress("0x3c499c542cef5e3811e1192ce70d8cc03d5c3359"): 6,
	common.HexToAddress("0xd6df932a45c0f255f85145f286ea0b292b21c90b"): 18,
	common.HexToAddress("0xb33eaad8d922b1083446dc23f610c2567fb5180f"): 18,
	common.HexToAddress("0x61299774020da444af134c82fa83e3810b309991"): 18,
	common.HexToAddress("0xc3c7d422809852031b44ab29eec9f1eff2a58756"): 18,
	common.HexToAddress("0xa3fa99a148fa48d14ed51d610c367c61876997f1"): 18,
	common.HexToAddress("0x385eeac5cb85a38a9a07a70c73e0a3271cfb54a7"): 18,
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("../page")))
	rawConf, err := os.ReadFile(os.Args[1])
	fmt.Println(string(rawConf))
	if err != nil {
		panic(err)
	}
	var conf Configuration
	json.Unmarshal(rawConf, &conf)
	wsrpcclients := make(map[string]*rpc.Client)
	for _, url := range conf.WsRpcs {
		go func(_url string) {
			client, err := rpc.Dial(_url)
			if err == nil {
				wsrpcclients[_url] = client
			} else {
				fmt.Println("error:", err, _url)
			}
		}(url)
	}

	http.HandleFunc("/swap", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		tokenIn := common.HexToAddress(r.URL.Query().Get("tokenIn"))
		amountInStr := r.URL.Query().Get("amountIn")
		tokenOut := common.HexToAddress(r.URL.Query().Get("tokenOut"))
		amountIn, err := strconv.ParseFloat(amountInStr, 64)
		if err != nil {
			http.Error(w, "Invalid amount", http.StatusBadRequest)
			return
		}
		var tInIx int64
		var tOutIx int64
		for _, t := range conf.Tokens {
			if t.Token.Cmp(tokenIn) == 0 {
				break
			}
			tInIx++
		}
		for _, t := range conf.Tokens {
			if t.Token.Cmp(tokenOut) == 0 {
				break
			}
			tOutIx++
		}
		amInF := big.NewFloat(amountIn)
		amIn := new(big.Int)
		amInF.Mul(amInF, new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(tokenDecimals[tokenIn])), nil))).Int(amIn)

		var response map[string]interface{}
		for _, rpcclient := range wsrpcclients {
			call2, err := new(caller.Batch).AddBlockByNumber("latest").AddFindRoutesForSingleToken(conf.Tokens, amIn, big.NewInt(tInIx), conf.Router, "latest").Execute(rpcclient)
			if err == nil {
				if call2[0] != nil && call2[1] != nil {
					block := call2[0].(map[string]interface{})
					baseFeeHex, _ := block["baseFeePerGas"].(string)
					if baseFeeHex == "" {
						return
					}
					baseFee, _ := hexutil.DecodeBig(baseFeeHex)
					gasPrice := new(big.Int).Add(new(big.Int).Mul(baseFee, big.NewInt(10)).Div(baseFee, big.NewInt(5)), big.NewInt(35e9))
					routes := call2[1].([]caller.Route)
					r := new(big.Float).SetInt(routes[tOutIx].AmOut)
					decDivisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(tokenDecimals[tokenOut])), nil))
					r.Quo(r, decDivisor)
					rr, _ := r.Float64()
					response = map[string]interface{}{"success": true, "tx": map[string]interface{}{"to": conf.Caller, "input": hexutil.Encode(routes[tOutIx].Calls), "gas": utils.RouteGas(routes[tOutIx].Calls), "gasPrice": gasPrice.Uint64()}, "amountOut": rr}
					break
				}
			} else {
				//+log.Println("Error:", err)
				response = map[string]interface{}{"success": false, "message": err}
			}
		}
		json.NewEncoder(w).Encode(response)
	})
	log.Println("Starting webserver on:", "http://127.0.0.1:8080/page")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
