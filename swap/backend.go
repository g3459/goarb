package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/g3459/goarb/caller"
	"github.com/g3459/goarb/contracts/bytecodes"
	"github.com/g3459/goarb/contracts/interfaces"
	"github.com/g3459/goarb/utils"
)

type Configuration struct {
	PrivateKey  *common.Hash       `json:"privateKey"`
	PoolFinder  *common.Address    `json:"poolFinder"`
	Caller      *common.Address    `json:"caller"`
	Tokens      []caller.TokenInfo `json:"tokens"`
	HttpRpcs    []string           `json:"httpRpcs"`
	WsRpcs      []string           `json:"wsRpcs"`
	ChainId     *big.Int           `json:"chainId"`
	MinMinerTip *big.Int           `json:"minMinerTip"`
}

func main() {
	rawConf, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var conf Configuration
	json.Unmarshal(rawConf, &conf)
	log.Println(conf)
	_privateKey, err := crypto.ToECDSA(conf.PrivateKey[:])
	if err != nil {
		panic(err)
	}
	sender := crypto.PubkeyToAddress(_privateKey.PublicKey)
	alloc := make(map[common.Address]types.Account)
	alloc[sender] = types.Account{Balance: new(big.Int).SetBytes(common.MaxHash[:])}
	simconf := func(nodeConf *node.Config, ethConf *ethconfig.Config) {
		ethConf.RPCEVMTimeout = 2 * time.Second
		ethConf.RPCGasCap = 0x7fffffffffffffff
		ethConf.Genesis.GasLimit = 0x7fffffffffffffff
		ethConf.Miner.GasCeil = 0x7fffffffffffffff
	}
	sim := simulated.NewBackend(alloc, simconf)
	simClient := sim.Client()
	auth, err := bind.NewKeyedTransactorWithChainID(_privateKey, big.NewInt(1337))
	if err != nil {
		log.Fatal(err)
	}
	router, _, _, err := bind.DeployContract(auth, interfaces.RouterABI, bytecodes.RouterBytecode, simClient)
	if err != nil {
		log.Fatal(err)
	}
	sim.Commit()
	http.Handle("/", http.FileServer(http.Dir("./")))
	fmt.Println(string(rawConf))
	if err != nil {
		panic(err)
	}

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
		var response *map[string]interface{}
		defer func() { json.NewEncoder(w).Encode(response) }()
		w.Header().Set("Access-Control-Allow-Origin", "*")
		tokenIn := common.HexToAddress(r.URL.Query().Get("tokenIn"))
		tokenOut := common.HexToAddress(r.URL.Query().Get("tokenOut"))
		amIn, _ := new(big.Int).SetString(r.URL.Query().Get("amountIn"), 10)
		var tInx int64
		var tOutx int64
		for _, t := range conf.Tokens {
			// fmt.Println(t.Token, tokenIn)
			if t.Token.Cmp(tokenIn) == 0 {
				break
			}
			tInx++
		}
		for _, t := range conf.Tokens {
			// fmt.Println(t.Token, tokenOut)
			if t.Token.Cmp(tokenOut) == 0 {
				break
			}
			tOutx++
		}
		log.Println("\n{\n    TokenIn: ", tokenIn, "\n    TokenOut: ", tokenOut, "\n    AmIn: ", amIn, "\n}")
		ethIn := new(big.Int).Mul(amIn, conf.Tokens[tInx].EthPX64)
		ethIn.Rsh(ethIn, 64)
		callch := make(chan interface{}, len(wsrpcclients))
		for _, rpcclient := range wsrpcclients {
			go func(_rpcclient *rpc.Client) {
				call, err := new(caller.Batch).BlockByNumber("latest", nil).FindPools(conf.Tokens, ethIn, conf.PoolFinder, "latest").Submit(_rpcclient)
				if err != nil {
					callch <- err
				} else {
					callch <- call
				}
			}(rpcclient)
		}
		for res := range callch {
			call, b := res.([]interface{})
			if !b {
				response = &map[string]interface{}{"success": false, "message": fmt.Sprint("BatchCallError:", res.(error))}
				log.Println(response)
				continue
			}
			block, b := call[0].(map[string]interface{})
			if !b {
				response = &map[string]interface{}{"success": false, "message": fmt.Sprint("Err:Block:", call[0].(error))}
				log.Println(response)
				continue
			}
			pools, b := call[1].([][][]*big.Int)
			if !b {
				response = &map[string]interface{}{"success": false, "message": fmt.Sprint("Err: Pools: ", call[1].(error))}
				log.Println(response)
				continue
			}
			baseFeeHex, _ := block["baseFeePerGas"].(string)
			baseFee, _ := hexutil.DecodeBig(baseFeeHex)
			minGasPrice := new(big.Int).Add(baseFee, conf.MinMinerTip)
			data, _ := interfaces.RouterABI.Pack("findRoutes", conf.Tokens, pools, big.NewInt(0), amIn, big.NewInt(tInx))
			msg := ethereum.CallMsg{
				From:     sender,
				To:       &router,
				GasPrice: minGasPrice,
				Data:     data,
			}
			log.Println("start", amIn, tInx)
			raw, err := simClient.CallContract(context.Background(), msg, nil)
			if err != nil {
				response = &map[string]interface{}{"success": false, "message": fmt.Sprint("Err: Route:", err)}
				log.Println(response)
				continue
			}
			res, _ := interfaces.RouterABI.Unpack("findRoutes", raw)
			log.Println("end", amIn, tInx)
			routes := res[0].([]struct {
				AmOut *big.Int "json:\"amOut\""
				Calls []uint8  "json:\"calls\""
			})
			response = &map[string]interface{}{"success": true, "tx": map[string]interface{}{"to": conf.Caller, "input": hexutil.Encode(routes[tOutx].Calls), "gas": utils.RouteGas(routes[tOutx].Calls), "gasPrice": minGasPrice.Uint64()}, "amountOut": routes[tOutx].AmOut}
			log.Println(response)
			return
		}
	})
	log.Println("Starting webserver on:", "http://127.0.0.1:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
