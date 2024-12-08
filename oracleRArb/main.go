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
	routerBytecode, _ = hex.DecodeString("60808060405234601957610d1f908161001e823930815050f35b5f80fdfe60806040526004361015610011575f80fd5b5f3560e01c633818b99d14610024575f80fd5b60803660031901126101b3576100386101bc565b6100406101cc565b906064359167ffffffffffffffff83116101b357366023840112156101b357826004013561006d81610212565b9361007b60405195866101f0565b8185526024602086019260051b820101903682116101b35760248101925b8284106100c4576100c06100b188604435888a610365565b6040939193519384938461025d565b0390f35b833567ffffffffffffffff81116101b3578201366043820112156101b35760248101356100f081610212565b916100fe60405193846101f0565b818352602060248185019360051b83010101903682116101b35760448101925b828410610138575050509082525060209384019301610099565b833567ffffffffffffffff81116101b35760249083010136603f820112156101b35760208101359167ffffffffffffffff83116101b757604051610186601f8501601f1916602001826101f0565b83815236604084860101116101b3575f60208581966040839701838601378301015281520193019261011e565b5f80fd5b6101dc565b6004359060ff821682036101b357565b6024359060ff821682036101b357565b634e487b7160e01b5f52604160045260245ffd5b90601f8019910116810190811067ffffffffffffffff8211176101b757604052565b67ffffffffffffffff81116101b75760051b60200190565b90602080835192838152019201905f5b8181106102475750505090565b825184526020938401939092019160010161023a565b9392906102729060608652606086019061022a565b938085036020820152825180865260208601906020808260051b8901019501915f905b8282106102b657505050506102b3939450604081840391015261022a565b90565b9091929560208080600193601f198d8203018652818b518051918291828552018484015e5f828201840152601f01601f1916010198019493919091019101610295565b9061030382610212565b61031060405191826101f0565b8281528092610321601f1991610212565b0190602036910137565b634e487b7160e01b5f52603260045260245ffd5b80511561034c5760200190565b61032b565b805182101561034c5760209160051b010190565b9193909361038060ff61037886516102f9565b961686610351565b5282519261038d84610212565b9361039b60405195866101f0565b8085526103aa601f1991610212565b015f5b8181106106b75750508360ff6103c383516102f9565b9460051b16905f198351610100031c805b6103de5750505050565b9194935f9791939697505f925b87518410156106a5576001841b908181161561069857189061040d8484610351565b51158015610684575b610677575f915b88518310156106645782851461065b575f610438868b610351565b5151151580610643575b156105f45750600161045e84610458888d610351565b51610351565b515b846105df578561048e895f935b85856104878d5f1961047f828a610351565b510195610351565b519361071e565b926104998884610351565b518211156105d1578b6104ac8b82610351565b516104bc60ff60d81b871661095d565b01913a8302903a6104cd8c85610351565b510290806105b3575b506104e18b87610351565b510390840313156105a457610520896105649561051a60019a9996610577999661050e8561052698610351565b525f1901928392610351565b52610975565b60a01b90565b90866b7fffffffff0000000000000160a01b0316179061059a575b61057261054e8a8c610351565b5191604051938491602083019190602083019252565b03601f1981018452836101f0565b6106fa565b6105818589610351565b5261058c8488610351565b5081841b17925b019161041d565b8460ff1b17610541565b50505050505091600190610593565b6105c3816105cb939488026106ca565b9286026106ca565b5f6104d6565b505050505091600190610593565b8561048e896105ed8361033f565b519361046d565b6105fe848b610351565b515115158061062b575b156106215761061b86610458868d610351565b51610460565b5091600190610593565b5061063a86610458868d610351565b51511515610608565b5061065284610458888d610351565b51511515610442565b91600190610593565b92989150926001905b01929097916103eb565b926001909891929861066d565b50866106908587610351565b515114610416565b929891936001915061066d565b919796949590949390925090806103d4565b60606020828801810191909152016103ad565b81156106d4570490565b634e487b7160e01b5f52601260045260245ffd5b805191908290602001825e015f815290565b61071c90610564610716949360405195869360208501906106e8565b906106e8565b565b929493925f908180805b895182101561095157818a016040810151959061074e6001600160a01b0388168b61099b565b6109455760200151936fffffffffffffffffffffffffffffffff6107728660801c90565b951692838a1561093c575b5061079561078e61078e8a60a01c90565b61ffff1690565b620f4240038088028b6107b3620f42408a02928881850191026106ca565b98868a111561084b5760ff60d81b8c1691908b8b8a8f8615801561092f575b610871575b5050505050506107e73a9161095d565b02958b8061085e575b508689039285870384131561084b5761082692889261081e926108138e60011b90565b0280920191026106ca565b039160011b90565b1261083c575050506040909294915b0190610728565b93919650935060409150610835565b5050505093919650935060409150610835565b61086a91978a026106ca565b958b6107f0565b806108ab6108916108a361089c61089161078e61078e6108b09860c81c90565b62ffffff1660020b90565b9360b01c90565b62ffffff1690565b6109d4565b9095156109075750926108dd926108cf6108d7936108e4960360801b90565b9101906106ca565b92610a17565b6002900a90565b115b6108f5578c5f8b8b8a8f6107d7565b50505093919650935060409150610835565b945050506108d76109206108dd92610929940160801b90565b8d8c03906106ca565b106108e6565b50600160d91b87146107d2565b9593505f61077d565b50945090604090610835565b97965050505094505050565b600160d91b0361096e57620493e090565b620186a090565b5f905b8065ffffffffffff81160361098e5760081b1790565b906008019060081c610978565b9060205b825181116109cd57828101516001600160a01b038381169116146109c55760200161099f565b505050600190565b5050505f90565b8190818082075f8312169105030290810160020b90620d89e7198160020b125f14610a025750620d89e71991565b91620d89e88213610a0f57565b620d89e89150565b60020b8060ff1d8181011890600182167001fffcb933bd6fad37aa2d162d1a59400102600160801b189160028116610d03575b60048116610ce7575b60088116610ccb575b60108116610caf575b60208116610c93575b60408116610c77575b60808116610c5b575b6101008116610c3f575b6102008116610c23575b6104008116610c07575b6108008116610beb575b6110008116610bcf575b6120008116610bb3575b6140008116610b97575b6180008116610b7b575b620100008116610b5f575b620200008116610b44575b620400008116610b29575b6208000016610b10575b5f12610b08575b60401c90565b5f1904610b02565b6b048a170391f7dc42444e8fa290910260801c90610afb565b6d2216e584f5fa1ea926041bedfe9890920260801c91610af1565b916e5d6af8dedb81196699c329225ee6040260801c91610ae6565b916f09aa508b5b7a84e1c677de54f3e99bc90260801c91610adb565b916f31be135f97d08fd981231505542fcfa60260801c91610ad0565b916f70d869a156d2a1b890bb3df62baf32f70260801c91610ac6565b916fa9f746462d870fdf8a65dc1f90e061e50260801c91610abc565b916fd097f3bdfd2022b8845ad8f792aa58250260801c91610ab2565b916fe7159475a2c29b7443b29c7fa6e889d90260801c91610aa8565b916ff3392b0822b70005940c7a398e4b70f30260801c91610a9e565b916ff987a7253ac413176f2b074cf7815e540260801c91610a94565b916ffcbe86c7900a88aedcffc83b479aa3a40260801c91610a8a565b916ffe5dee046a99a2a811c461f1969c30530260801c91610a80565b916fff2ea16466c96a3843ec78b326b528610260801c91610a77565b916fff973b41fa98c081472e6896dfb254c00260801c91610a6e565b916fffcb9843d60f6159c9db58835c9266440260801c91610a65565b916fffe5caca7e10e4e61c3624eaa0941cd00260801c91610a5c565b916ffff2e50f5f656932ef12357cf3c7fdcc0260801c91610a53565b916ffff97272373d413259a46990580e213a0260801c91610a4a56")
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
			// continue
			Log(4, "START", number, sts2.Sub(sts))
			go func() {
				minGasPrice := new(big.Int).Add(baseFee, conf.MinMinerTip)
				var txCalls []byte
				var txGasLimit uint64
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
				for gasPrice.Cmp(minGasPrice) >= 0 && txCalls == nil {
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
									txGas := route.GasUsage
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
									padded := make([]byte, 16)
									copy(padded[16-len(amIn.Bytes()):], amIn.Bytes())
									txCalls = append(route.Calls, padded...)
									txGasLimit = route.GasUsage.Uint64() + 60000
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
				if txCalls != nil {
					Log(1, txCalls, callsGasPriceLimit, number)
					if !conf.FakeBalance {
						lastCalls = txCalls
						minerTip := new(big.Int).Sub(callsGasPriceLimit, baseFee)
						if minerTip.Cmp(conf.MaxMinerTip) > 0 {
							minerTip = conf.MaxMinerTip
						}
						if callsGasPriceLimit.Cmp(conf.MaxGasPrice) > 0 {
							callsGasPriceLimit = conf.MaxGasPrice
						}
						b := new(caller.Batch).SendTx(&types.DynamicFeeTx{ChainID: conf.ChainId, Nonce: nonce, GasTipCap: minerTip, GasFeeCap: callsGasPriceLimit, Gas: txGasLimit, To: conf.Caller, Value: new(big.Int), Data: txCalls, AccessList: AccessListForCalls(txCalls)}, conf.PrivateKey, nil)
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
