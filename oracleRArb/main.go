package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
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
	MinGasBen   *big.Int          `json:"minGasBen"`
	MinRatio    float64           `json:"minRatio"`
	Protocols   []caller.Protocol `json:"protocols"`
	FakeBalance bool              `json:"fakeBalance"`
	LogLevel    int               `json:"logLevel"`
	//RouteDepth  uint8             `json:"routeDepth"`
	RouteMaxLen uint8 `json:"routeMaxLen"`
	//LogFile     string            `json:"logFile"`
	//Timeout     time.Duration     `json:"timeout"`
	Polling       time.Duration `json:"polling"`
	ExecTime      time.Duration `json:"execTime"`
	IsOpRollup    bool          `json:"isOpRollup"`
	MaxL1GasPrice *big.Int      `json:"maxL1GasPrice"`
}

var (
	conf              Configuration
	router            *common.Address
	ethPriceX64Oracle []*big.Int
	rpcClients        = make(map[string]*rpc.Client)
	rpcClientsBanMap  = map[*rpc.Client]time.Time{}
	simClient         *rpc.Client
	hBlockn           uint64
	hNonce            uint64
	lastTxNonce       uint64
	sender            *common.Address
	//logFile           *os.File
	lastCalls         []byte
	routerBytecode, _ = hex.DecodeString("60808060405234601957610d20908161001e823930815050f35b5f80fdfe60806040526004361015610011575f80fd5b5f3560e01c633818b99d14610024575f80fd5b60803660031901126101b3576100386101bc565b6100406101cc565b906064359167ffffffffffffffff83116101b357366023840112156101b357826004013561006d81610212565b9361007b60405195866101f0565b8185526024602086019260051b820101903682116101b35760248101925b8284106100c4576100c06100b188604435888a610365565b6040939193519384938461025d565b0390f35b833567ffffffffffffffff81116101b3578201366043820112156101b35760248101356100f081610212565b916100fe60405193846101f0565b818352602060248185019360051b83010101903682116101b35760448101925b828410610138575050509082525060209384019301610099565b833567ffffffffffffffff81116101b35760249083010136603f820112156101b35760208101359167ffffffffffffffff83116101b757604051610186601f8501601f1916602001826101f0565b83815236604084860101116101b3575f60208581966040839701838601378301015281520193019261011e565b5f80fd5b6101dc565b6004359060ff821682036101b357565b6024359060ff821682036101b357565b634e487b7160e01b5f52604160045260245ffd5b90601f8019910116810190811067ffffffffffffffff8211176101b757604052565b67ffffffffffffffff81116101b75760051b60200190565b90602080835192838152019201905f5b8181106102475750505090565b825184526020938401939092019160010161023a565b9392906102729060608652606086019061022a565b938085036020820152825180865260208601906020808260051b8901019501915f905b8282106102b657505050506102b3939450604081840391015261022a565b90565b9091929560208080600193601f198d8203018652818b518051918291828552018484015e5f828201840152601f01601f1916010198019493919091019101610295565b9061030382610212565b61031060405191826101f0565b8281528092610321601f1991610212565b0190602036910137565b634e487b7160e01b5f52603260045260245ffd5b80511561034c5760200190565b61032b565b805182101561034c5760209160051b010190565b9193909361038060ff61037886516102f9565b961686610351565b5282519261038d84610212565b9361039b60405195866101f0565b8085526103aa601f1991610212565b015f5b8181106106b85750508360ff6103c383516102f9565b9460051b16905f198351610100031c805b6103de5750505050565b9194935f9791939697505f925b87518410156106a6576001841b908181161561069957189061040d8484610351565b51158015610685575b610678575f915b88518310156106655782851461065c575f610438868b610351565b5151151580610644575b156105f55750600161045e84610458888d610351565b51610351565b515b846105e0578561048f895f935b85856104888d600119610480828a610351565b510195610351565b519361071f565b9261049a8884610351565b518211156105d2578b6104ad8b82610351565b516104bd60ff60d81b871661095e565b01913a8302903a6104ce8c85610351565b510290806105b4575b506104e28b87610351565b510390840313156105a557610521896105659561051b60019a9996610578999661050f8561052798610351565b528b1901928392610351565b52610976565b60a01b90565b90866b7fffffffff0000000000000160a01b0316179061059b575b61057361054f8a8c610351565b5191604051938491602083019190602083019252565b03601f1981018452836101f0565b6106fb565b6105828589610351565b5261058d8488610351565b5081841b17925b019161041d565b8460ff1b17610542565b50505050505091600190610594565b6105c4816105cc939488026106cb565b9286026106cb565b5f6104d7565b505050505091600190610594565b8561048f896105ee8361033f565b519361046d565b6105ff848b610351565b515115158061062c575b156106225761061c86610458868d610351565b51610460565b5091600190610594565b5061063b86610458868d610351565b51511515610609565b5061065384610458888d610351565b51511515610442565b91600190610594565b92989150926001905b01929097916103eb565b926001909891929861066e565b50866106918587610351565b515114610416565b929891936001915061066e565b919796949590949390925090806103d4565b60606020828801810191909152016103ad565b81156106d5570490565b634e487b7160e01b5f52601260045260245ffd5b805191908290602001825e015f815290565b61071d90610565610717949360405195869360208501906106e9565b906106e9565b565b929493925f908180805b895182101561095257818a016040810151959061074f6001600160a01b0388168b61099c565b6109465760200151936fffffffffffffffffffffffffffffffff6107738660801c90565b951692838a1561093d575b5061079661078f61078f8a60a01c90565b61ffff1690565b620f4240038088028b6107b4620f42408a02928881850191026106cb565b98868a111561084c5760ff60d81b8c1691908b8b8a8f86158015610930575b610872575b5050505050506107e83a9161095e565b02958b8061085f575b508689039285870384131561084c5761082792889261081f926108148e60011b90565b0280920191026106cb565b039160011b90565b1261083d575050506040909294915b0190610729565b93919650935060409150610836565b5050505093919650935060409150610836565b61086b91978a026106cb565b958b6107f1565b806108ac6108926108a461089d61089261078f61078f6108b19860c81c90565b62ffffff1660020b90565b9360b01c90565b62ffffff1690565b6109d5565b9095156109085750926108de926108d06108d8936108e5960360801b90565b9101906106cb565b92610a18565b6002900a90565b115b6108f6578c5f8b8b8a8f6107d8565b50505093919650935060409150610836565b945050506108d86109216108de9261092a940160801b90565b8d8c03906106cb565b106108e7565b50600160d91b87146107d3565b9593505f61077e565b50945090604090610836565b97965050505094505050565b600160d91b0361096f57620493e090565b620186a090565b5f905b8065ffffffffffff81160361098f5760081b1790565b906008019060081c610979565b9060205b825181116109ce57828101516001600160a01b038381169116146109c6576020016109a0565b505050600190565b5050505f90565b8190818082075f8312169105030290810160020b90620d89e7198160020b125f14610a035750620d89e71991565b91620d89e88213610a1057565b620d89e89150565b60020b8060ff1d8181011890600182167001fffcb933bd6fad37aa2d162d1a59400102600160801b189160028116610d04575b60048116610ce8575b60088116610ccc575b60108116610cb0575b60208116610c94575b60408116610c78575b60808116610c5c575b6101008116610c40575b6102008116610c24575b6104008116610c08575b6108008116610bec575b6110008116610bd0575b6120008116610bb4575b6140008116610b98575b6180008116610b7c575b620100008116610b60575b620200008116610b45575b620400008116610b2a575b6208000016610b11575b5f12610b09575b60401c90565b5f1904610b03565b6b048a170391f7dc42444e8fa290910260801c90610afc565b6d2216e584f5fa1ea926041bedfe9890920260801c91610af2565b916e5d6af8dedb81196699c329225ee6040260801c91610ae7565b916f09aa508b5b7a84e1c677de54f3e99bc90260801c91610adc565b916f31be135f97d08fd981231505542fcfa60260801c91610ad1565b916f70d869a156d2a1b890bb3df62baf32f70260801c91610ac7565b916fa9f746462d870fdf8a65dc1f90e061e50260801c91610abd565b916fd097f3bdfd2022b8845ad8f792aa58250260801c91610ab3565b916fe7159475a2c29b7443b29c7fa6e889d90260801c91610aa9565b916ff3392b0822b70005940c7a398e4b70f30260801c91610a9f565b916ff987a7253ac413176f2b074cf7815e540260801c91610a95565b916ffcbe86c7900a88aedcffc83b479aa3a40260801c91610a8b565b916ffe5dee046a99a2a811c461f1969c30530260801c91610a81565b916fff2ea16466c96a3843ec78b326b528610260801c91610a78565b916fff973b41fa98c081472e6896dfb254c00260801c91610a6f565b916fffcb9843d60f6159c9db58835c9266440260801c91610a66565b916fffe5caca7e10e4e61c3624eaa0941cd00260801c91610a5d565b916ffff2e50f5f656932ef12357cf3c7fdcc0260801c91610a54565b916ffff97272373d413259a46990580e213a0260801c91610a4b56")
)

func main() {
	startConf()
	ExecTime(conf.ExecTime * time.Second)
	///
	// if len(conf.LogFile) > 0 {
	// 	logFile, err = os.OpenFile(conf.LogFile, os.O_APPEND|os.O_WRONLY, 0600)
	// 	if os.IsNotExist(err) {
	// 		logFile, err = os.OpenFile(conf.LogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	// 		if err != nil {
	// 			Log(-1, "OpenLogFile Err: ", err)
	// 		}
	// 		if _, err = logFile.WriteString("number,tokenIn,tokenOut,amountIn,amountOut,ethIn,ethOut,benefit(eth),gasPriceLimit,gasPrice,gasLimit,nonce\n"); err != nil {
	// 			Log(-1, "WriteLogFile Err: ", err)
	// 		}
	// 	}
	// 	defer logFile.Close()
	// }
	startUsdOracles()
	startRpcClients(conf.RpcUrls)

	var err error
	batch := caller.Batch{}
	number := uint64(0)
	pools := make([][][]byte, len(conf.TokenConfs))
	tokens := make([]common.Address, len(conf.TokenConfs))
	for i, v := range conf.TokenConfs {
		tokens[i] = *v.Token
	}
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
	var l1GasPrice *big.Int
	if conf.IsOpRollup {
		batch = batch.L1GasPrice(func(res interface{}) {
			_l1GasPrice, b := res.(*big.Int)
			if !b {
				err = errors.New("L1GasPrice Err: " + res.(error).Error())
				return
			}
			l1GasPrice = _l1GasPrice
		})
	}
	var minGasPrice *big.Int
	batch = batch.GasPrice(func(res interface{}) {
		_gasPrice, b := res.(*big.Int)
		if !b {
			err = errors.New("GasPrice Err: " + res.(error).Error())
			return
		}
		minGasPrice = _gasPrice
	})
	nonce := uint64(0)
	batch = batch.Nonce(sender, "latest", func(res interface{}) {
		var b bool
		nonce, b = res.(uint64)
		if !b {
			err = errors.New("Nonce Err: " + res.(error).Error())
			return
		}
	})

	for {
		for _, rpcclient := range rpcClients {
			if clientBanned(rpcclient) {
				continue
			}
			dlt := time.Now().Add(conf.Polling * time.Millisecond)
			deadline, cancel := context.WithDeadline(context.Background(), dlt)
			deadline, cancel = context.WithCancel(deadline)
			go func() {
				sts := time.Now()
				err = nil
				_, err2 := batch.FindPoolsCheckBlockNumber(conf.MinLiqEth, tokens, conf.Protocols, hBlockn+1, conf.PoolFinder, "pending", func(res interface{}) {
					_res, b := res.([]interface{})
					if !b {
						err = errors.New("FindPools Err: " + res.(error).Error())
						return
					}
					pools = _res[0].([][][]byte)
					number = _res[1].(uint64)
				}).Submit(deadline, rpcclient)
				if err2 != nil {
					banClient(rpcclient, conf.Polling*time.Millisecond*30)
					Log(2, "BatchRPC Err: ", err2)
					cancel()
					return
				}
				if err != nil {
					banClient(rpcclient, conf.Polling*time.Millisecond*30)
					Log(2, "BatchExec Err: ", err)
					cancel()
					return
				}
				if nonce < hNonce {
					Log(3, fmt.Sprintf("nonce(%v) < hNonce(%v)", nonce, hNonce))
					cancel()
					return
				}
				// if nonce <= lastTxNonce {
				// 	Log(3, fmt.Sprintf("nonce(%v) <= lastTxNonce(%v)", nonce, lastTxNonce))
				// 	cancel()
				// 	return
				// }
				if number < hBlockn || len(pools) == 0 {
					Log(3, fmt.Sprintf("number(%v) < hBlockn(%v)", number, hBlockn), len(pools))
					cancel()
					return
				}
				if minGasPrice.Cmp(conf.MaxGasPrice) > 0 {
					Log(3, fmt.Sprintf("blockMinGasPrice(%v) > confMaxGasPrice(%v)", minGasPrice, conf.MaxGasPrice))
					return
				}
				if conf.IsOpRollup && l1GasPrice.Cmp(conf.MaxL1GasPrice) > 0 {
					Log(3, fmt.Sprintf("l1GasPrice(%v) > confMaxL1GasPrice(%v)", l1GasPrice, conf.MaxL1GasPrice))
					return
				}
				// if number == hBlockn && nonce == hNonce {
				// 	Log(3, "number == hBlockn && nonce == hNonce")
				// 	cancel()
				// 	return
				// }
				Log(3, fmt.Sprintf("\nNEW_BATCH {GasPrice:%v, Block:%v, Nonce:%v, ResTime:%v}", minGasPrice, number, nonce, time.Since(sts)))
				defer func(t time.Time) { Log(3, fmt.Sprintf("END_BATCH %v\n", time.Since(t))) }(time.Now())
				if number > hBlockn {
					hBlockn = number
				}
				if nonce > hNonce {
					hNonce = nonce
				}
				// token := common.HexToAddress("0x2791bca1f2de4661ed88a30c99a7a9449aa84174")
				// token := common.HexToAddress("0x7ceB23fD6bC0adD59E62ac25578270cFf1b9f619")
				// token := common.HexToAddress("0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270")

				// token := common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913")
				// token := common.HexToAddress("0x4200000000000000000000000000000000000006")
				// token := common.HexToAddress("0x0b2c639c533813f4aa9d7837caf62653d097ff85")
				// token := common.HexToAddress("0x7f5c764cbc14f9669b88837ca1490cca17c31607")
				// token := common.HexToAddress("0x4200000000000000000000000000000000000042")
				// res, errr := caller.Batch{}.ExecuteApprove(conf.Caller, &token, sender, common.MaxHash.Big(), big.NewInt(100), conf.MaxGasPrice, nonce, conf.ChainId, conf.PrivateKey, nil).Submit(context.Background(), rpcclient)
				// Log(0, res, errr)
				// return
				var txCalls []byte
				var txGasLimit uint64
				var checkFuncs []func() = make([]func(), 0)
				callsGasPriceLimit := new(big.Int).Set(minGasPrice)
				gasPrice := new(big.Int).Lsh(conf.MaxGasPrice, 2)
				for i := range conf.TokenConfs {
					if amounts[i] == nil {
						continue
					}
					Log(4, "Token:", conf.TokenConfs[i].Token, ", AmIn:", amounts[i], ", Price:", ethPriceX64Oracle[i])
				}
				var wg sync.WaitGroup
				Log(4, "START_COMP")
				sts2 := time.Now()
				for gasPrice.Cmp(minGasPrice) >= 0 {
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
								f := func() {
									for tOutx, route := range routes {
										// ll := 0
										// if len(pools[tInx]) > 0 {
										// 	ll += len(pools[tInx][tOutx]) / 0x40
										// }
										// if len(pools[tOutx]) > 0 {
										// 	ll += len(pools[tOutx][tInx]) / 0x40
										// }
										// fmt.Println(tInx, tOutx, amIn, route.AmOut, len(route.Calls)/0x20, ll, route.GasUsage)
										// continue
										if ethPriceX64Oracle[tOutx] == nil {
											continue
										}
										if len(route.Calls) == 0 {
											Log(5, tInx, tOutx, amIn, gasPrice, "noCalls")
											continue
										}
										if bytes.Equal(route.Calls, lastCalls) {
											Log(5, tInx, tOutx, amIn, gasPrice, "calls==lastCalls", lastCalls)
											continue
										}
										ethOut := new(big.Int).Mul(route.AmOut, ethPriceX64Oracle[tOutx])
										ethOut.Rsh(ethOut, 64)
										ben := new(big.Int).Sub(ethOut, ethIn)
										if conf.IsOpRollup {
											l1Fees := big.NewInt(int64(16*(len(route.Calls)+int((amIn.BitLen()+7)/8)) + 1088))
											l1Fees.Mul(l1Fees, l1GasPrice)
											ben.Sub(ben, l1Fees)
										}
										if ben.Sign() < 0 {
											Log(5, tInx, tOutx, amIn, gasPrice, fmt.Sprintf("ethIn(%vwei)-ethOut(%vwei)<0", ethIn, ethOut))
											continue
										}
										ratiotemp := new(big.Float).SetInt(ethOut)
										ratiotemp.Quo(ratiotemp, new(big.Float).SetInt(ethIn))
										ratio, _ := ratiotemp.Float64()
										if ratio < conf.MinRatio {
											Log(5, tInx, tOutx, amIn, gasPrice, fmt.Sprintf("ratio(%v)<MinRatio(%v)", ratio, conf.MinRatio))
											continue
										}
										txGas := new(big.Int).Set(route.GasUsage)
										gasFees := new(big.Int).Mul(txGas, gasPrice)
										if new(big.Int).Sub(ben, gasFees).Sign() > 0 {
											Log(4, tInx, tOutx, amIn, gasPrice, fmt.Sprintf("ben(%vwei)-gasFees(%vwei)>0", ben, gasFees))
											continue
										}
										txGas.Add(txGas, conf.MinGasBen)
										gasPriceLimit := new(big.Int).Div(ben, txGas)
										if gasPriceLimit.Cmp(callsGasPriceLimit) < 0 {
											Log(4, tInx, tOutx, amIn, gasPrice, fmt.Sprintf("gasPriceLimit(%v)<callsGasPriceLimit(%v)", gasPriceLimit, callsGasPriceLimit))
											continue
										}
										if gasPriceLimit.Cmp(conf.MaxGasPrice) > 0 {
											Log(4, tInx, tOutx, amIn, gasPrice, fmt.Sprintf("gasPriceLimit(%v)>MaxGasPrice(%v)", gasPriceLimit, conf.MaxGasPrice))
											continue
										}
										Log(4, tInx, tOutx, amIn, gasPrice, route.Calls, gasPriceLimit)
										callsGasPriceLimit = gasPriceLimit
										txCalls = append(route.Calls, amIn.Bytes()...)
										txGasLimit = route.GasUsage.Uint64() + 150000
									}
								}
								checkFuncs = append(checkFuncs, f)
							}(amIn, uint8(i), gasPrice)
							// Log(5, "START", i, amIn, gasPrice)
							amIn = new(big.Int).Rsh(amIn, 1)
						}
					}
					gasPrice = new(big.Int).Rsh(gasPrice, 1)
				}
				wg.Wait()
				Log(4, "END_COMP", time.Since(sts2))
				for time.Now().Compare(dlt) < 0 {
					for _, f := range checkFuncs {
						f()
					}
					if txCalls != nil {
						Log(1, txCalls, callsGasPriceLimit, number)
						if !conf.FakeBalance {
							lastTxNonce = nonce
							lastCalls = txCalls
							if callsGasPriceLimit.Cmp(conf.MaxGasPrice) > 0 {
								callsGasPriceLimit.Set(conf.MaxGasPrice)
							}
							b := new(caller.Batch).SendTx(&types.DynamicFeeTx{ChainID: conf.ChainId, Nonce: nonce, GasTipCap: callsGasPriceLimit, GasFeeCap: callsGasPriceLimit, Gas: txGasLimit, To: conf.Caller, Value: new(big.Int), Data: txCalls, AccessList: AccessListForCalls(txCalls)}, conf.PrivateKey, nil)
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
									Log(3, *hash, number)
								}(rpcclient)
							}
						}
						return
					}
					<-time.After(time.Millisecond * 100)
				}
			}()
			<-deadline.Done()
			cancel()
		}
	}
}

func startConf() {
	rawConf, err := os.ReadFile(os.Args[1])
	if err != nil {
		Log(-1, "ReadConfFile Err: ", err)
	}
	err = json.Unmarshal(rawConf, &conf)
	if err != nil {
		Log(-1, "UnmarshalConfFile Err: ", err)
	}
	logLevel = conf.LogLevel
	_sender := crypto.PubkeyToAddress(crypto.ToECDSAUnsafe((conf.PrivateKey)[:]).PublicKey)
	sender = &_sender
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
			deadline, cancel := context.WithDeadline(context.Background(), time.Now().Add(5000*time.Millisecond))
			defer cancel()
			client, err := rpc.DialContext(deadline, url)
			if err != nil {
				Log(1, "rpcDial Err: ", err, url)
				return
			}
			batch := caller.Batch{}
			res, err := batch.BlockNumber(nil).Submit(deadline, client)
			if err != nil {
				Log(1, "rpcBlockByNumber Err: ", err, url)
				return
			}
			number, b := res[0].(uint64)
			if !b {
				Log(1, "rpcBlockByNumber Err: ", err, url)
				return
			}
			Log(1, "rpcBlockByNumber: ", url, number)
			mu.Lock()
			rpcClients[url] = client
			mu.Unlock()
		}()
	}
	wg.Wait()
	if len(rpcClients) == 0 {
		Log(-1, "Unable to connect any rpc")
	}
}

var banmu sync.Mutex

func banClient(client *rpc.Client, d time.Duration) {
	banmu.Lock()
	rpcClientsBanMap[client] = time.Now().Add(d)
	banmu.Unlock()
}

func clientBanned(client *rpc.Client) bool {
	banmu.Lock()
	b := time.Now().Compare(rpcClientsBanMap[client]) < 0
	banmu.Unlock()
	return b
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
