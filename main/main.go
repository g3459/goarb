package main

import (
	//"context"

	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"net/http"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/g3459/goarb/caller"
	"github.com/g3459/goarb/utils"
)

type Configuration struct {
	PrivateKey string   `json:"privateKey"`
	Tokens     []string `json:"tokens"`
	WsRpcs     []string `json:"wsRpcs"`
	ChainId    uint     `json:"chainId"`
}

type ChainInfo struct {
	router string
	caller string
}

var chainsInfo = map[uint]ChainInfo{
	137: {
		router: "0x1f016f951602f0E6aeA12174cC520E8E9a7f4DC0",
		caller: "0x2213F075c552Aa26A5eeD66775973f96741beFAa",
	},
}

type TokenInfo struct {
	name     string
	eth      float64
	decimals uint8
}

var tokensInfo = map[string]TokenInfo{
	"0xc2132d05d31c914a87c6611c10748aeb04b58e8f": {"USD", 1.4, 6},
	"0x2791bca1f2de4661ed88a30c99a7a9449aa84174": {"USD", 1.4, 6},
	"0x0d500b1d8e8ef31e21c99d1db9a6444d3adf1270": {"MATIC", 1, 18},
	"0x7ceb23fd6bc0add59e62ac25578270cff1b9f619": {"ETH", 5300, 18},
	"0x8f3cf7ad23cd3cadbd9735aff958023239c6a063": {"USD", 1.4, 18},
	"0x1bfd67037b42cf73acf2047067bd4f2c47d9bfd6": {"BTC", 97000, 8},
	"0x53e0bca35ec356bd5dddfebbd1fc0fd03fabad39": {"LINK", 24, 18},
	"0x3c499c542cef5e3811e1192ce70d8cc03d5c3359": {"USD", 1.4, 6},
	"0xd6df932a45c0f255f85145f286ea0b292b21c90b": {"AAVE", 140, 18},
	"0xb33eaad8d922b1083446dc23f610c2567fb5180f": {"UNI", 13, 18},
	"0x61299774020da444af134c82fa83e3810b309991": {"RNDR", 14, 18},
	"0xc3c7d422809852031b44ab29eec9f1eff2a58756": {"LDO", 3.2, 18},
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
		client, err := rpc.Dial(url)
		if err == nil {
			wsrpcclients[url] = client
		} else {
			fmt.Println("error:", err, url)
		}
	}
	tokenList := make([]common.Address, len(conf.Tokens))
	ethPricesX64 := make([]*big.Int, len(conf.Tokens))
	for i, t := range conf.Tokens {
		tokenList[i] = common.HexToAddress(t)
		ethPricesX64[i] = big.NewInt(int64(tokensInfo[t].eth * (1 << 32)))
		ethPricesX64[i].Mul(ethPricesX64[i], big.NewInt(1<<32))
		ethPricesX64[i].Mul(ethPricesX64[i], big.NewInt(int64(math.Pow10(int(18-tokensInfo[t].decimals)))))
	}
	http.HandleFunc("/swap", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		tokenIn := r.URL.Query().Get("tokenIn")
		amountInStr := r.URL.Query().Get("amountIn")
		tokenOut := r.URL.Query().Get("tokenOut")
		amountIn, err := strconv.ParseFloat(amountInStr, 64)
		if err != nil {
			http.Error(w, "Invalid amount", http.StatusBadRequest)
			return
		}
		var tInIx int64
		var tOutIx int64
		for _, t := range conf.Tokens {
			if t == tokenIn {
				break
			}
			tInIx++
		}
		for _, t := range conf.Tokens {
			if t == tokenOut {
				break
			}
			tOutIx++
		}
		amInF := big.NewFloat(amountIn)
		amIn := new(big.Int)
		amInF.Mul(amInF, new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(tokensInfo[tokenIn].decimals)), nil))).Int(amIn)

		var response map[string]interface{}
		for _, rpcclient := range wsrpcclients {
			call2, err := new(caller.Batch).AddBlockByNumber("latest").AddFindRoutesForSingleToken(tokenList, ethPricesX64[tInIx], amIn, big.NewInt(tInIx), chainsInfo[conf.ChainId].caller, chainsInfo[conf.ChainId].router, "latest").Execute(rpcclient)
			if err == nil {
				if call2[0] != nil && call2[1] != nil {
					block := call2[0].(map[string]interface{})
					baseFeeHex, _ := block["baseFeePerGas"].(string)
					if baseFeeHex == "" {
						return
					}
					baseFee := new(big.Int).SetBytes(utils.HexToBytes(baseFeeHex))
					gasPrice := new(big.Int).Add(baseFee, big.NewInt(30e9))
					routes := call2[1].([]caller.Route)
					fmt.Println(routes)
					r := new(big.Float).SetInt(routes[tOutIx].AmOut)
					decDivisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(tokensInfo[tokenOut].decimals)), nil))
					r.Quo(r, decDivisor)
					rr, _ := r.Float64()
					response = map[string]interface{}{"success": true, "tx": map[string]interface{}{"to": chainsInfo[conf.ChainId].caller, "input": utils.BytesToHex(routes[tOutIx].Calls), "gas": 1000000, "gasPrice": gasPrice.Mul(gasPrice, big.NewInt(2)).Uint64()}, "amountOut": rr}
					break
				}
			} else {
				log.Println("Error:", err)
				response = map[string]interface{}{"success": false, "message": err}
			}
		}
		json.NewEncoder(w).Encode(response)
	})
	log.Println("Starting webserver on:", "http://127.0.0.1:8080/page")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
