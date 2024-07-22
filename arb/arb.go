package main

import (
	//"context"

	"bytes"
	"context"
	"encoding/json"
	"log"
	"math/big"
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
	MaxGasPrice *big.Int           `json:"maxGasPrice"`
	MinEth      *big.Int           `json:"minEth"`
	MaxMinerTip *big.Int           `json:"maxMinerTip"`
	MinMinerTip *big.Int           `json:"minMinerTip"`
	MinGasBen   *big.Int           `json:"minGasBen"`
}

var conf Configuration
var router common.Address
var wsrpcclients = make(map[string]*rpc.Client)
var httprpcclients = make(map[string]*rpc.Client)
var hNumber uint64
var sender common.Address
var lastCalls []byte
var simClient simulated.Client

func main() {
	rawConf, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	json.Unmarshal(rawConf, &conf)
	log.Println(conf)
	_privateKey, err := crypto.ToECDSA(conf.PrivateKey[:])
	if err != nil {
		panic(err)
	}
	sender = crypto.PubkeyToAddress(_privateKey.PublicKey)
	alloc := make(map[common.Address]types.Account)
	alloc[sender] = types.Account{Balance: new(big.Int).SetBytes(common.MaxHash[:])}
	simconf := func(nodeConf *node.Config, ethConf *ethconfig.Config) {
		ethConf.RPCEVMTimeout = 2 * time.Second
		ethConf.RPCGasCap = 0x7fffffffffffffff
		ethConf.Genesis.GasLimit = 0x7fffffffffffffff
		ethConf.Miner.GasCeil = 0x7fffffffffffffff
	}
	sim := simulated.NewBackend(alloc, simconf)
	simClient = sim.Client()
	auth, err := bind.NewKeyedTransactorWithChainID(_privateKey, big.NewInt(1337))
	if err != nil {
		log.Fatal(err)
	}
	router, _, _, err = bind.DeployContract(auth, interfaces.RouterABI, bytecodes.RouterBytecode, simClient)
	if err != nil {
		log.Fatal(err)
	}
	sim.Commit()
	// for _, url := range conf.HttpRpcs {
	// 	go func(_url string) {
	// 		client, err := rpc.Dial(_url)
	// 		if err == nil {
	// 			httprpcclients[_url] = client
	// 		} else {
	// 			log.Println("error:", err, _url)
	// 		}
	// 		httprpcclients[_url] = client
	// 	}(url)
	// }
	for _, url := range conf.WsRpcs {
		go func(_url string) {
			client, err := rpc.Dial(_url)
			if err == nil {
				defer client.Close()
				// token := common.HexToAddress("0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174")
				// amIn := new(big.Int).Mul(big.NewInt(43974800), big.NewInt(1))
				// nonce := uint64(18748)
				// call, err := new(caller.Batch).AddExecuteTransfer(conf.Caller, &token, &sender, amIn, big.NewInt(30e9), conf.MaxGasPrice, nonce, conf.ChainId, conf.PrivateKey).Execute(client)
				// log.Println(call, err)
				// return
				wsrpcclients[_url] = client
				newHeadsCh := make(chan map[string]interface{}, len(wsrpcclients))
				defer close(newHeadsCh)
				client.EthSubscribe(context.Background(), newHeadsCh, "newHeads")
				for block := range newHeadsCh {
					go headsCallback(block, client)
				}
			} else {
				log.Println("wserror:", err, _url)
			}
		}(url)
	}
	<-time.After(10 * time.Minute)
	os.Exit(0)
}

func headsCallback(block map[string]interface{}, _rpcclient *rpc.Client) {
	blockNumberHex, _ := block["number"].(string)
	number, _ := hexutil.DecodeUint64(blockNumberHex)
	if number <= hNumber {
		return
	}
	log.Println("Number:", number)
	batch := caller.Batch{}
	for _, v := range conf.Tokens {
		batch = batch.AddCallBalanceOf(v.Token, conf.Caller, blockNumberHex)
	}
	call, err := batch.AddCallFindPools(conf.Tokens, conf.MinEth, conf.PoolFinder, blockNumberHex).AddNonce(&sender, blockNumberHex).Execute(_rpcclient)
	if number <= hNumber {
		return
	}
	if err != nil {
		log.Println("BatchCallError:", err)
		return
	}
	nonce, b := call[len(call)-1].(uint64)
	if !b {
		log.Println("Err: -1", call[len(call)-1].(error))
		return
	}
	pools, b := call[len(call)-2].([][][]*big.Int)
	if !b {
		log.Println("Err: -2", call[len(call)-2].(error))
		return
	}
	hNumber = number
	baseFeeHex, _ := block["baseFeePerGas"].(string)
	baseFee, _ := hexutil.DecodeBig(baseFeeHex)
	minGasPrice := new(big.Int).Add(baseFee, conf.MinMinerTip)
	// routeCall, err := new(caller.Batch).AddCallFindRoutes(conf.Tokens, pools, conf.MinEth, big.NewInt(40000000), big.NewInt(0), minGasPrice, router, "latest").Execute(simClient)
	for i, v := range conf.Tokens {
		amIn, b := call[i].(*big.Int)
		if !b {
			log.Println("Err: -3", call[i].(error))
			return
		}
		ethIn := new(big.Int).Mul(amIn, v.EthPX64)
		ethIn.Rsh(ethIn, 64)
		for ethIn.Cmp(conf.MinEth) > 0 {
			go func(_amIn *big.Int, _tInx int) {
				data, _ := interfaces.RouterABI.Pack("findRoutes", conf.Tokens, pools, big.NewInt(0), _amIn, big.NewInt(int64(_tInx)))
				msg := ethereum.CallMsg{
					From:     sender,
					To:       &router,
					GasPrice: minGasPrice,
					Data:     data,
				}
				log.Println("start", _amIn, _tInx, number)
				raw, err := simClient.CallContract(context.Background(), msg, nil)
				if err != nil {
					log.Println("------------------------------->RouteErr: ", err)
					return
				}
				res, _ := interfaces.RouterABI.Unpack("findRoutes", raw)
				log.Println("end", _amIn, _tInx, number)
				routes := res[0].([]struct {
					AmOut *big.Int "json:\"amOut\""
					Calls []uint8  "json:\"calls\""
				})
				hRouteGasPrice := new(big.Int)
				var route *caller.Route
				for tOutx := range routes {
					_route := caller.Route(routes[tOutx])
					//fmt.Println("\n{\ntInx:", _tInx, "\ntOutx:", tOutx, "\namIn:", _amIn, "\namOut:", _route.AmOut, "\nCalls:", _route.Calls, "\n}\n")

					if _tInx == tOutx && len(_route.Calls) > 0 {
						ethInX64 := new(big.Int).Mul(_amIn, conf.Tokens[_tInx].EthPX64)
						ethOutX64 := new(big.Int).Mul(_route.AmOut, conf.Tokens[tOutx].EthPX64)
						txGas := big.NewInt(int64(utils.RouteGas(_route.Calls)))
						txFeeX64 := new(big.Int).Mul(txGas, minGasPrice)
						txFeeX64.Lsh(txFeeX64, 64)
						ethOutX64.Add(ethOutX64, txFeeX64)
						ben := new(big.Int).Sub(ethOutX64, ethInX64)
						ben.Rsh(ben, 64)
						txGas.Add(txGas, conf.MinGasBen)
						routeGasPrice := new(big.Int).Div(ben, txGas)
						routeGasPrice.Add(routeGasPrice, minGasPrice)
						if routeGasPrice.Cmp(hRouteGasPrice) > 0 {
							hRouteGasPrice = routeGasPrice
							route = &_route
						}
					}
				}
				if hRouteGasPrice.Cmp(minGasPrice) <= 0 {
					return
				}
				if bytes.Equal(route.Calls, lastCalls) {
					return
				}
				lastCalls = route.Calls
				if hRouteGasPrice.Cmp(conf.MaxGasPrice) > 0 {
					hRouteGasPrice = conf.MaxGasPrice
				}
				minerTip := new(big.Int).Sub(hRouteGasPrice, baseFee)
				if minerTip.Cmp(conf.MaxMinerTip) > 0 {
					minerTip = conf.MaxMinerTip
				}
				for url, rpcclient := range wsrpcclients {
					go func(_url string, _rpcclient *rpc.Client) {
						res, err := new(caller.Batch).AddExecuteRoute(route.Calls, nonce, conf.Caller, minerTip, hRouteGasPrice, conf.ChainId, conf.PrivateKey).Execute(_rpcclient)
						log.Println("{\nHash/Err: ", res, err, "\nNonce: ", nonce, "\nBlock:", number, "\nAmIn", _amIn, "\nRoute:", route, "\nLastCalls:", lastCalls, "\n}")
					}(url, rpcclient)
				}
				for url, rpcclient := range httprpcclients {
					go func(_url string, _rpcclient *rpc.Client) {
						res, err := new(caller.Batch).AddExecuteRoute(route.Calls, nonce, conf.Caller, minerTip, hRouteGasPrice, conf.ChainId, conf.PrivateKey).Execute(_rpcclient)
						log.Println("{\nHash/Err: ", res, err, "\nNonce: ", nonce, "\nBlock:", number, "\nAmIn", _amIn, "\nRoute:", route, "\nLastCalls:", lastCalls, "\n}")
					}(url, rpcclient)
				}
			}(amIn, i)
			amIn = new(big.Int).Rsh(amIn, 1)
			ethIn = new(big.Int).Rsh(ethIn, 1)
		}
	}
	// return

}
