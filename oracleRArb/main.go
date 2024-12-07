package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"math/big"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/g3459/goarb/caller"
	"github.com/g3459/goarb/simulated"
	"github.com/gorilla/websocket"
)

type Oracle struct {
	Name     string `json:"name"`
	Decimals int    `json:"decimals"`
	Active   bool   `json:"active"`
}

type TokenConf struct {
	Token       *common.Address `json:"token"`
	Oracle      Oracle          `json:"oracle"`
	FakeBalance *big.Int        `json:"fakeBalance"`
}

type Configuration struct {
	PrivateKey  *common.Hash      `json:"privateKey"`
	PoolFinder  *common.Address   `json:"poolFinder"`
	Caller      *common.Address   `json:"caller"`
	TokenConfs  []TokenConf       `json:"tokens"`
	RpcUrls     []string          `json:"rpcUrls"`
	ChainId     *big.Int          `json:"chainId"`
	MaxGasPrice *big.Int          `json:"maxGasPrice"`
	MinEth      *big.Int          `json:"minEth"`
	MinLiqEth   *big.Int          `json:"minLiqEth"`
	MaxMinerTip *big.Int          `json:"maxMinerTip"`
	MinMinerTip *big.Int          `json:"minMinerTip"`
	MinGasBen   *big.Int          `json:"minGasBen"`
	MinRatio    float64           `json:"minRatio"`
	Protocols   []caller.Protocol `json:"protocols"`
	FakeBalance bool              `json:"fakeBalance"`
	LogLevel    int               `json:"logLevel"`
	RouteDepth  uint8             `json:"routeDepth"`
	RouteMaxLen uint8             `json:"routeMaxLen"`
	LogFile     string            `json:"logFile"`
	//Timeout     time.Duration     `json:"timeout"`
	Polling     time.Duration `json:"polling"`
	ExecTimeout time.Duration `json:"execTimeout"`
}

var (
	conf              Configuration
	router            *common.Address
	ethPriceX64Oracle []*big.Int
	rpcClients        = make(map[string]*rpc.Client)
	rpcClientsBanMap  = map[*rpc.Client]time.Time{}
	simClient         *rpc.Client
	hNumber           uint64
	sender            *common.Address
	logFile           *os.File
	lastCalls         []byte
	routerBytecode, _ = hex.DecodeString("60808060405234601957610d08908161001e823930815050f35b5f80fdfe60806040526004361015610011575f80fd5b5f3560e01c633818b99d14610024575f80fd5b60803660031901126101b2576100386101bb565b6100406101cb565b906064359067ffffffffffffffff82116101b257366023830112156101b25781600401359261006e84610211565b9261007c60405194856101ef565b8484526024602085019560051b820101903682116101b25760248101955b8287106100c3576100af866044358688610343565b906100bf60405192839283610229565b0390f35b863567ffffffffffffffff81116101b2578201366043820112156101b25760248101356100ef81610211565b916100fd60405193846101ef565b818352602060248185019360051b83010101903682116101b25760448101925b82841061013757505050908252506020968701960161009a565b833567ffffffffffffffff81116101b25760249083010136603f820112156101b25760208101359167ffffffffffffffff83116101b657604051610185601f8501601f1916602001826101ef565b83815236604084860101116101b2575f60208581966040839701838601378301015281520193019261011d565b5f80fd5b6101db565b6004359060ff821682036101b257565b6024359060ff821682036101b257565b634e487b7160e01b5f52604160045260245ffd5b90601f8019910116810190811067ffffffffffffffff8211176101b657604052565b67ffffffffffffffff81116101b65760051b60200190565b604081016040825282518091526020606083019301905f5b8181106102c1575050506020818303910152815180825260208201916020808360051b8301019401925f915b83831061027c57505050505090565b909192939460208080600193601f19868203018752818a518051918291828552018484015e5f828201840152601f01601f19160101970195949190910192019061026d565b8251855260209485019490920191600101610241565b906102e182610211565b6102ee60405191826101ef565b82815280926102ff601f1991610211565b0190602036910137565b634e487b7160e01b5f52603260045260245ffd5b80511561032a5760200190565b610309565b805182101561032a5760209160051b010190565b9290939161035e60ff61035684516102d7565b96168661032f565b5280519061036b82610211565b9161037960405193846101ef565b808352610388601f1991610211565b015f5b8181106106a257505060ff829460051b166103a682516102d7565b925f198351610100031c805b6103bd575050505050565b91945f979491939697505f925b8751841015610691576001841b90818116156106845718906103ec848761032f565b51158015610670575b610663575f915b885183101561065057828514610647575f610417868b61032f565b515115158061062f575b156105e05750600161043d84610437888d61032f565b5161032f565b515b846105ca57868961046e8a5f945b8686610467875f1961045f828a61032f565b51019561032f565b5193610707565b93610479898461032f565b518211156105bb578961048c858261032f565b5161049c60ff60d81b8816610946565b01913a8302903a6104ad8d8561032f565b5102908061059d575b506104c18c8761032f565b5103908403131561058d5761050361054d9560019998958c6104fd6105609a9761050f976104f2846105099961032f565b525f1901918361032f565b5261032f565b5161095e565b60a01b90565b90866b7fffffffff0000000000000160a01b03161790610583575b61055b6105378a8c61032f565b5191604051938491602083019190602083019252565b03601f1981018452836101ef565b6106e3565b61056a858961032f565b52610575848861032f565b5081841b17925b01916103fc565b8460ff1b1761052a565b505050505050509160019061057c565b6105ad816105b5939488026106b3565b9286026106b3565b5f6104b6565b5050505050509160019061057c565b868961046e8a6105d98361031d565b519461044d565b6105ea848b61032f565b5151151580610617575b1561060d5761060786610437868d61032f565b5161043f565b509160019061057c565b5061062686610437868d61032f565b515115156105f4565b5061063e84610437888d61032f565b51511515610421565b9160019061057c565b92989150926001905b01929097916103ca565b9260019098919298610659565b508661067c858761032f565b5151146103f5565b9298919360019150610659565b8095989794929196935090946103b2565b80606060208093870101520161038b565b81156106bd570490565b634e487b7160e01b5f52601260045260245ffd5b805191908290602001825e015f815290565b6107059061054d6106ff949360405195869360208501906106d1565b906106d1565b565b929493925f908180805b895182101561093a57818a01604081015195906107376001600160a01b0388168b610984565b61092e5760200151936fffffffffffffffffffffffffffffffff61075b8660801c90565b951692838a15610925575b5061077e6107776107778a60a01c90565b61ffff1690565b620f4240038088028b61079c620f42408a02928881850191026106b3565b98868a11156108345760ff60d81b8c1691908b8b8a8f86158015610918575b61085a575b5050505050506107d03a91610946565b02958b80610847575b50868903928587038413156108345761080f928892610807926107fc8e60011b90565b0280920191026106b3565b039160011b90565b12610825575050506040909294915b0190610711565b9391965093506040915061081e565b505050509391965093506040915061081e565b61085391978a026106b3565b958b6107d9565b8061089461087a61088c61088561087a6107776107776108999860c81c90565b62ffffff1660020b90565b9360b01c90565b62ffffff1690565b6109bd565b9095156108f05750926108c6926108b86108c0936108cd960360801b90565b9101906106b3565b92610a00565b6002900a90565b115b6108de578c5f8b8b8a8f6107c0565b5050509391965093506040915061081e565b945050506108c06109096108c692610912940160801b90565b8d8c03906106b3565b106108cf565b50600160d91b87146107bb565b9593505f610766565b5094509060409061081e565b97965050505094505050565b600160d91b0361095757620493e090565b620186a090565b5f905b8065ffffffffffff8116036109775760081b1790565b906008019060081c610961565b9060205b825181116109b657828101516001600160a01b038381169116146109ae57602001610988565b505050600190565b5050505f90565b8190818082075f8312169105030290810160020b90620d89e7198160020b125f146109eb5750620d89e71991565b91620d89e882136109f857565b620d89e89150565b60020b8060ff1d8181011890600182167001fffcb933bd6fad37aa2d162d1a59400102600160801b189160028116610cec575b60048116610cd0575b60088116610cb4575b60108116610c98575b60208116610c7c575b60408116610c60575b60808116610c44575b6101008116610c28575b6102008116610c0c575b6104008116610bf0575b6108008116610bd4575b6110008116610bb8575b6120008116610b9c575b6140008116610b80575b6180008116610b64575b620100008116610b48575b620200008116610b2d575b620400008116610b12575b6208000016610af9575b5f12610af1575b60401c90565b5f1904610aeb565b6b048a170391f7dc42444e8fa290910260801c90610ae4565b6d2216e584f5fa1ea926041bedfe9890920260801c91610ada565b916e5d6af8dedb81196699c329225ee6040260801c91610acf565b916f09aa508b5b7a84e1c677de54f3e99bc90260801c91610ac4565b916f31be135f97d08fd981231505542fcfa60260801c91610ab9565b916f70d869a156d2a1b890bb3df62baf32f70260801c91610aaf565b916fa9f746462d870fdf8a65dc1f90e061e50260801c91610aa5565b916fd097f3bdfd2022b8845ad8f792aa58250260801c91610a9b565b916fe7159475a2c29b7443b29c7fa6e889d90260801c91610a91565b916ff3392b0822b70005940c7a398e4b70f30260801c91610a87565b916ff987a7253ac413176f2b074cf7815e540260801c91610a7d565b916ffcbe86c7900a88aedcffc83b479aa3a40260801c91610a73565b916ffe5dee046a99a2a811c461f1969c30530260801c91610a69565b916fff2ea16466c96a3843ec78b326b528610260801c91610a60565b916fff973b41fa98c081472e6896dfb254c00260801c91610a57565b916fffcb9843d60f6159c9db58835c9266440260801c91610a4e565b916fffe5caca7e10e4e61c3624eaa0941cd00260801c91610a45565b916ffff2e50f5f656932ef12357cf3c7fdcc0260801c91610a3c565b916ffff97272373d413259a46990580e213a0260801c91610a3356")
)

func main() {
	conf = readConf()
	ExecTimeout(conf.ExecTimeout * time.Second)
	///
	// if len(conf.LogFile) > 0 {
	// 	logFile, err = os.OpenFile(conf.LogFile, os.O_APPEND|os.O_WRONLY, 0600)
	// 	if os.IsNotExist(err) {
	// 		logFile, err = os.OpenFile(conf.LogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	// 		if err != nil {
	// 			Log(-1, "OpenLogFile Err: ", err)
	// 		}
	// 		if _, err = logFile.WriteString("number,tokenIn,tokenOut,amountIn,amountOut,ethIn,ethOut,benefit(eth),gasPriceLimit,minGasPrice,gasLimit,nonce\n"); err != nil {
	// 			Log(-1, "WriteLogFile Err: ", err)
	// 		}
	// 	}
	// 	defer logFile.Close()
	// }
	startUsdOracles()
	startRpcClients(conf.RpcUrls)
	//batch declaration
	var err error
	batch := caller.Batch{}
	amounts := make([]*big.Int, len(conf.TokenConfs))
	if !conf.FakeBalance {
		for i, v := range conf.TokenConfs {
			if !v.Oracle.Active {
				continue
			}
			batch = batch.BalanceOf(v.Token, conf.Caller, "pending", func(res interface{}) {
				am, b := res.(*big.Int)
				if !b {
					err = errors.New("BalanceOf " + v.Token.Hex() + " Err: " + res.(error).Error())
					return
				}
				amounts[i] = am
			})
		}
	} else {
		for i, v := range conf.TokenConfs {
			if !v.Oracle.Active {
				continue
			}
			if v.FakeBalance != nil {
				amounts[i] = v.FakeBalance
			} else {
				amounts[i] = new(big.Int)
			}
		}
	}
	var number uint64
	var baseFee *big.Int
	blockInfo := map[string]interface{}{}
	batch = batch.BlockByNumber("pending", func(res interface{}) {
		var b bool
		_blockInfo, b := res.(*map[string]interface{})
		if !b {
			err = errors.New("Block Err: " + res.(error).Error())
			return
		}
		blockInfo = *_blockInfo
		number, _ = hexutil.DecodeUint64((blockInfo)["number"].(string))
		baseFee, _ = hexutil.DecodeBig(blockInfo["baseFeePerGas"].(string))
	})
	pools := make([][][]byte, len(conf.TokenConfs))
	tokens := make([]common.Address, len(conf.TokenConfs))
	for i, v := range conf.TokenConfs {
		tokens[i] = *v.Token
	}
	batch = batch.FindPools(conf.MinLiqEth, tokens, conf.Protocols, conf.PoolFinder, "pending", func(res interface{}) {
		var b bool
		pools, b = res.([][][]byte)
		if !b {
			err = errors.New("FindPools Err: " + res.(error).Error())
			return
		}
	})
	nonce := uint64(0)
	batch = batch.Nonce(sender, "pending", func(res interface{}) {
		var b bool
		nonce, b = res.(uint64)
		if !b {
			err = errors.New("Nonce Err: " + res.(error).Error())
			return
		}
	})
	///
	//logic execution
	for {
		for _, rpcclient := range rpcClients {
			if clientBanned(rpcclient) {
				continue
			}
			err = nil
			deadline, cancel := context.WithDeadline(context.Background(), time.Now().Add(conf.Polling*time.Millisecond))
			sts := time.Now()
			_, err2 := batch.Submit(deadline, rpcclient)
			if err2 != nil {
				banClient(rpcclient, conf.Polling*time.Millisecond*40)
				Log(2, "BatchRPC Err: ", err2)
				continue
			}
			if err != nil {
				banClient(rpcclient, conf.Polling*time.Millisecond*40)
				Log(2, "BatchExec Err: ", err)
				continue
			}
			sts2 := time.Now()
			if number < hNumber {
				continue
			}
			if number > hNumber {
				hNumber = number
			}
			// token := common.HexToAddress("0x2791bca1f2de4661ed88a30c99a7a9449aa84174")
			// token := common.HexToAddress("0x7ceB23fD6bC0adD59E62ac25578270cFf1b9f619")
			// token := common.HexToAddress("0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270")
			// res, errr := caller.Batch{}.ExecuteApprove(conf.Caller, &token, sender, common.MaxHash.Big(), conf.MinMinerTip, conf.MaxGasPrice, nonce, conf.ChainId, conf.PrivateKey, nil).Submit(context.Background(), rpcclient)
			// Log(0, res, errr)
			continue
			Log(4, "START", number, sts2.Sub(sts))
			go func() {
				minGasPrice := new(big.Int).Add(baseFee, conf.MinMinerTip)
				var calls []byte
				callsGasPriceLimit := minGasPrice
				gasPrice := new(big.Int).Lsh(conf.MaxGasPrice, 2)
				for i := range conf.TokenConfs {
					if amounts[i] == nil {
						continue
					}
					Log(4, "Token:", conf.TokenConfs[i].Token, ", AmIn:", amounts[i], ", Price:", ethPriceX64Oracle[i])
				}
				var wg sync.WaitGroup
				var mu sync.Mutex
				for gasPrice.Cmp(minGasPrice) >= 0 && calls == nil {
					for i := range conf.TokenConfs {
						if amounts[i] == nil {
							continue
						}
						if ethPriceX64Oracle[i] == nil {
							continue
						}
						amInMin := new(big.Int).Div(new(big.Int).Lsh(conf.MinEth, 64), ethPriceX64Oracle[i])
						amIn := new(big.Int).Set(amounts[i])
						for amIn.Cmp(amInMin) > 0 {
							wg.Add(1)
							go func(amIn *big.Int, tInx uint8, gasPrice *big.Int) {
								defer wg.Done()
								res, err := new(caller.Batch).FindRoutes(conf.RouteMaxLen, tInx, amIn, pools, gasPrice, router, "pending", nil).Submit(context.Background(), simClient)
								if err != nil {
									Log(2, "FindRoutesRPC Err: ", err)
									return
								}
								routes, b := res[0].([]caller.Route)
								if !b {
									Log(2, amIn, tInx, "FindRoutesExec Err: ", res[0].(error))
									return
								}
								ethIn := new(big.Int).Mul(amIn, ethPriceX64Oracle[tInx])
								ethIn.Rsh(ethIn, 64)
								mu.Lock()
								for tOutx, route := range routes {
									// ll := 0
									// if len(pools[tInx]) > 0 {
									// 	ll += len(pools[tInx][tOutx]) / 0x40
									// }
									// if len(pools[tOutx]) > 0 {
									// 	ll += len(pools[tOutx][tInx]) / 0x40
									// }
									// fmt.Println(tInx, tOutx, conf.TokenConfs[tOutx].Token, route.AmOut, len(route.Calls)/0x20, ll)
									// continue
									if ethPriceX64Oracle[tOutx] == nil || len(route.Calls) == 0 || bytes.Equal(route.Calls, lastCalls) {
										continue
									}
									ethOut := new(big.Int).Mul(route.AmOut, ethPriceX64Oracle[tOutx])
									ethOut.Rsh(ethOut, 64)
									ben := new(big.Int).Sub(ethOut, ethIn)
									if ben.Sign() < 0 {
										continue
									}
									ratiotemp := new(big.Float).SetInt(ethOut)
									ratiotemp.Quo(ratiotemp, new(big.Float).SetInt(ethIn))
									ratio, _ := ratiotemp.Float64()
									if ratio < conf.MinRatio {
										continue
									}
									txGas := big.NewInt(int64(CallsGas(route.Calls)))
									if new(big.Int).Sub(ben, new(big.Int).Mul(txGas, gasPrice)).Sign() > 0 {
										continue
									}
									txGas.Add(txGas, conf.MinGasBen)
									gasPriceLimit := new(big.Int).Div(ben, txGas)
									if gasPriceLimit.Cmp(conf.MaxGasPrice) > 0 {
										continue
									}
									if gasPriceLimit.Cmp(callsGasPriceLimit) < 0 {
										continue
									}
									callsGasPriceLimit = gasPriceLimit
									calls = route.Calls
								}
								mu.Unlock()
							}(amIn, uint8(i), gasPrice)
							Log(5, "START", i, amIn, gasPrice)
							amIn = new(big.Int).Rsh(amIn, 1)
						}
					}
					wg.Wait()
					gasPrice = new(big.Int).Rsh(gasPrice, 1)
				}
				ets := time.Now()
				Log(4, "END", number, ets.Sub(sts))
				if calls != nil {
					Log(1, calls, callsGasPriceLimit, number)
					if !conf.FakeBalance {
						lastCalls = calls
						minerTip := new(big.Int).Sub(callsGasPriceLimit, baseFee)
						if minerTip.Cmp(conf.MaxMinerTip) > 0 {
							minerTip = conf.MaxMinerTip
						}
						if callsGasPriceLimit.Cmp(conf.MaxGasPrice) > 0 {
							callsGasPriceLimit = conf.MaxGasPrice
						}
						b := new(caller.Batch).SendTx(&types.DynamicFeeTx{ChainID: conf.ChainId, Nonce: nonce, GasTipCap: minerTip, GasFeeCap: callsGasPriceLimit, Gas: ExecuteCallsGas(calls), To: conf.Caller, Value: new(big.Int), Data: calls, AccessList: AccessListForCalls(calls)}, conf.PrivateKey, nil)
						for _, rpcclient := range rpcClients {
							go func(rpcclient *rpc.Client) {
								res, err := b.Submit(context.Background(), rpcclient)
								if err != nil {
									Log(3, "ExecutePoolCallsRPC Err: ", err)
									return
								}
								hash, b := res[0].(*string)
								if !b {
									Log(3, "ExecutePoolCallsSend Err: ", res[0].(error))
									return
								}
								Log(3, *hash, number, ets.Sub(sts2))
							}(rpcclient)
						}
					}
				}
			}()
			<-deadline.Done()
			cancel()
		}
	}
	///
}

func readConf() (conf Configuration) {
	rawConf, err := os.ReadFile(os.Args[1])
	if err != nil {
		Log(-1, "ReadConfFile Err: ", err)
	}
	json.Unmarshal(rawConf, &conf)
	_sender := crypto.PubkeyToAddress(crypto.ToECDSAUnsafe((conf.PrivateKey)[:]).PublicKey)
	sender = &_sender
	return conf
}

func startRpcClients(rpcUrls []string) {
	var err error
	sim := simulated.NewSimulated()
	simClient = sim.Client().Client()
	// b := caller.Batch{}
	// res, err := b.SendTx(&types.DynamicFeeTx{ChainID: big.NewInt(1337), Nonce: 0, GasTipCap: new(big.Int), GasFeeCap: new(big.Int), Gas: 10000000, Value: new(big.Int), Data: bytecodes.RouterBytecode}, conf.PrivateKey, nil).Submit(context.Background(), simClient)
	// Log(0, *res[0].(*string))
	router, err = simulated.DeployContract(sim, routerBytecode)
	if err != nil {
		Log(-1, "simDeployContract Err: ", err)
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, url := range rpcUrls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			deadline, cancel := context.WithDeadline(context.Background(), time.Now().Add(3000*time.Millisecond))
			defer cancel()
			client, err := rpc.DialContext(deadline, url)
			if err != nil {
				Log(1, "rpcDial Err: ", err, url)
				return
			}
			batch := caller.Batch{}
			res, err := batch.BlockByNumber("pending", nil).Submit(deadline, client)
			if err != nil {
				Log(1, "rpcDial Err: ", err, url)
				return
			}
			_blockInfo, b := res[0].(*map[string]interface{})
			if !b {
				Log(1, "rpcDial Err: ", err, url)
				return
			}
			number, _ := hexutil.DecodeUint64((*_blockInfo)["number"].(string))
			mu.Lock()
			Log(1, "rpcDial: ", url, number)
			rpcClients[url] = client
			mu.Unlock()
		}()
	}
	wg.Wait()
	if len(rpcClients) == 0 {
		Log(-1, errors.New("Unable to connect any rpc"))
	}
}

func banClient(client *rpc.Client, d time.Duration) {
	rpcClientsBanMap[client] = time.Now().Add(d)
}

func clientBanned(client *rpc.Client) bool {
	return time.Now().Compare(rpcClientsBanMap[client]) < 0
}

func startUsdOracles() {
	ethPriceX64Oracle = make([]*big.Int, len(conf.TokenConfs))
	ethPriceX64Oracle[0] = new(big.Int).Lsh(big.NewInt(1), 64)
	var wg sync.WaitGroup
	for i, v := range conf.TokenConfs {
		if v.Oracle.Active && len(v.Oracle.Name) > 0 && v.Oracle.Name != "usd" || i == 0 {
			wg.Add(1)
			go func(baseToken string) {
				defer wg.Done()
				err := startBinanceUsdOracle(baseToken)
				if err != nil {
					Log(-1, "binanceDial Err: ", err)
				} else {
					Log(1, "binanceDial: ", baseToken)
					return
				}
				err = startBybitUsdOracle(baseToken)
				if err != nil {
					Log(-1, "bybitDial Err: ", err)
				} else {
					Log(1, "bybitDial: ", baseToken)
					return
				}
			}(v.Oracle.Name)
		}
	}
	wg.Wait()
}

func startBinanceUsdOracle(baseToken string) error {
	wsURL := "wss://fstream.binance.com/ws/" + strings.ToLower(baseToken) + "usdt@aggTrade"
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return err
	}
	go func() {
		defer c.Close()
		go func() {
			pingticker := time.NewTicker(30 * time.Second)
			defer pingticker.Stop()
			for range pingticker.C {
				time.Sleep(30 * time.Second)
				err := c.WriteMessage(websocket.PongMessage, nil)
				if err != nil {
					Log(-1, "Error sending pong:", err)
					return
				}
			}
		}()
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				Log(-1, "binanceRead Err:", err)
			}
			var res map[string]interface{}
			if err := json.Unmarshal(message, &res); err != nil {
				Log(-1, "binanceUnmarshal Err:", err)
			}
			price, err := strconv.ParseFloat(res["p"].(string), 64)
			if err != nil {
				Log(-1, "binanceParsePrice Err:", err)
			}
			updatePriceUsdOracle(baseToken, price)
		}
	}()
	return nil
}

func startBybitUsdOracle(baseToken string) error {
	wsURL := "wss://stream.bybit.com/v5/public/linear"
	var err error
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			c.Close()
		}
	}()
	err = c.WriteJSON(struct {
		Op   string   `json:"op"`
		Args []string `json:"args"`
	}{"subscribe", []string{"tickers." + strings.ToUpper(baseToken) + "USDT"}})
	if err != nil {
		return err
	}
	_, message, err := c.ReadMessage()
	if err != nil {
		return err
	}
	var res map[string]interface{}
	if err = json.Unmarshal(message, &res); err != nil {
		return err
	}
	if !res["success"].(bool) {
		return errors.New(res["ret_msg"].(string))
	}
	go func() {
		defer c.Close()
		go func() {
			pingticker := time.NewTicker(20 * time.Second)
			defer pingticker.Stop()
			for range pingticker.C {
				err := c.WriteJSON(struct {
					Op string `json:"op"`
				}{"ping"})
				if err != nil {
					Log(-1, "bybitWrite Err:", err)
				}
			}
		}()
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				Log(-1, "bybitRead Err:", err)
			}
			var res map[string]interface{}
			if err := json.Unmarshal(message, &res); err != nil {
				Log(-1, "bybitUnmarshal Err:", err)
			}
			if res["success"] != nil {
				if !res["success"].(bool) {
					Log(-1, "bybitMsg Err:", errors.New("tickers."+strings.ToUpper(baseToken)+"USDT "+res["ret_msg"].(string)))
				} else {
					Log(4, "bybitMsg:", "tickers."+strings.ToUpper(baseToken)+"USDT "+res["ret_msg"].(string))
				}
			} else {
				price, err := strconv.ParseFloat(res["data"].(map[string]interface{})["lastPrice"].(string), 64)
				if err != nil {
					Log(-1, "bybitParsePrice Err:", err)
				}
				updatePriceUsdOracle(baseToken, price)
			}
		}
	}()
	return nil
}

var usdEthPrice float64

func updatePriceUsdOracle(baseToken string, price float64) {
	if baseToken == conf.TokenConfs[0].Oracle.Name {
		usdEthPrice = 1 / price
		ethPX64 := ToX64Int(usdEthPrice)
		for i, v := range conf.TokenConfs {
			if v.Oracle.Active && v.Oracle.Name == "usd" {
				if ethPriceX64Oracle[i] == nil {
					ethPriceX64Oracle[i] = new(big.Int)
				}
				ethPriceX64Oracle[i].Mul(ethPX64, big.NewInt(int64(math.Pow10(18-v.Oracle.Decimals))))
			}
		}
	} else {
		if usdEthPrice > 0 {
			ethPX64 := ToX64Int(price * usdEthPrice)
			for i, v := range conf.TokenConfs {
				if v.Oracle.Active && v.Oracle.Name == baseToken {
					if ethPriceX64Oracle[i] == nil {
						ethPriceX64Oracle[i] = new(big.Int)
					}
					ethPriceX64Oracle[i].Mul(ethPX64, big.NewInt(int64(math.Pow10(18-v.Oracle.Decimals))))
				}
			}
		}
	}
}
