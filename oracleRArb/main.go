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
	routerBytecode, _ = hex.DecodeString("60c03461008057601f610ebb38819003918201601f19168301916001600160401b0383118484101761008457808492604094855283398101031261008057610052602061004b83610098565b9201610098565b60a052608052604051610e1590816100a68239608051816108b4015260a0518181816103c1015261079f0152f35b5f80fd5b634e487b7160e01b5f52604160045260245ffd5b519081151582036100805756fe60806040526004361015610011575f80fd5b5f3560e01c633818b99d14610024575f80fd5b346101b85760803660031901126101b85761003d6101c1565b6100456101d1565b906064359167ffffffffffffffff83116101b857366023840112156101b857826004013561007281610217565b9361008060405195866101f5565b8185526024602086019260051b820101903682116101b85760248101925b8284106100c9576100c56100b688604435888a61036a565b60409391935193849384610262565b0390f35b833567ffffffffffffffff81116101b8578201366043820112156101b85760248101356100f581610217565b9161010360405193846101f5565b818352602060248185019360051b83010101903682116101b85760448101925b82841061013d57505050908252506020938401930161009e565b833567ffffffffffffffff81116101b85760249083010136603f820112156101b85760208101359167ffffffffffffffff83116101bc5760405161018b601f8501601f1916602001826101f5565b83815236604084860101116101b8575f602085819660408397018386013783010152815201930192610123565b5f80fd5b6101e1565b6004359060ff821682036101b857565b6024359060ff821682036101b857565b634e487b7160e01b5f52604160045260245ffd5b90601f8019910116810190811067ffffffffffffffff8211176101bc57604052565b67ffffffffffffffff81116101bc5760051b60200190565b90602080835192838152019201905f5b81811061024c5750505090565b825184526020938401939092019160010161023f565b9392906102779060608652606086019061022f565b938085036020820152825180865260208601906020808260051b8901019501915f905b8282106102bb57505050506102b8939450604081840391015261022f565b90565b9091929560208080600193601f198d8203018652818b518051918291828552018484015e5f828201840152601f01601f191601019801949391909101910161029a565b9061030882610217565b61031560405191826101f5565b8281528092610326601f1991610217565b0190602036910137565b634e487b7160e01b5f52603260045260245ffd5b8051156103515760200190565b610330565b80518210156103515760209160051b010190565b93909360609261038760ff61037f87516102fe565b971687610356565b5283519061039482610217565b916103a260405193846101f5565b8083526103b1601f1991610217565b015f5b81811061072f57505081947f00000000000000000000000000000000000000000000000000000000000000009182610719575b60ff9060059493941b16915f198251610100031c805b610408575050505050565b929591945f9891949798505f935b8651851015610706576001851b90818116156106f95718906104388584610356565b511580156106e5575b6106d8575f915b87518310156106c5578286146106bc578688858c868a5f6104698287610356565b51511515806106a4575b156106505750946104976104c1926104916104ba9798600199610356565b51610356565b515b8961063c57868d5f9788915b6001196104b2828b610356565b510195610356565b5193610793565b9390916104ce8985610356565b5183111561060f57908893929161058d575b509261050f61050961054d946001989794610503610560988b1901928392610356565b52610a6b565b60a01b90565b90866b7fffffffff0000000000000160a01b03161790610583575b61055b6105378b8d610356565b5191604051938491602083019190602083019252565b03601f1981018452836101f5565b61076f565b61056a858a610356565b526105758489610356565b5081841b17925b0191610448565b8460ff1b1761052a565b909192508961059c8c82610356565b516105ac60ff60d81b8716610a53565b01913a8302903a6105bd8c85610356565b5102908061061e575b506105d18b87610356565b5103908403131561060f576105098961054d9561050360019a999661056099966105fe8561050f98610356565b5295985050949798509450506104e0565b5050505050509160019061057c565b61062e816106369394880261073f565b92860261073f565b5f6105c6565b868d61064787610344565b519788916104a5565b9592939490916106609082610356565b515115158061068c575b1561060f57916106866104c1926104918b6104ba989796610356565b51610499565b5061069b826104918b84610356565b5151151561066a565b506106b3836104918489610356565b51511515610473565b9160019061057c565b92999150936001905b0193909891610416565b93600190999192996106ce565b50876106f18688610356565b515114610441565b92999194600191506106ce565b91989792969095929490935090806103fd565b945060ff61072782516102fe565b9590506103e7565b80866020809387010152016103b4565b8115610749570490565b634e487b7160e01b5f52601260045260245ffd5b805191908290602001825e015f815290565b6107919061054d61078b9493604051958693602085019061075d565b9061075d565b565b949392945f5f915f905f7f0000000000000000000000000000000000000000000000000000000000000000935b8a51821015610a4657818b01604081015196906107e66001600160a01b0389168c610a91565b610a3a5760200151916fffffffffffffffffffffffffffffffff61080a8460801c90565b931691828b15610a31575b5061082d6108266108268b60a01c90565b61ffff1690565b620f4240038087028c61084b620f424088029287818501910261073f565b9685881115610a1f5760ff60d81b8d16918c918b8f85158015610a12575b610958575b5050505061088b575b50505050505060409095915b0190946107c0565b610899909994959699610a53565b3a02958c80610945575b50868a0392858703841315610932577f00000000000000000000000000000000000000000000000000000000000000006108ee575b5050505050505060409093905f80808080610877565b61091592889261090d926109028d60011b90565b02809201910261073f565b039160011b90565b1261092357808080806108d8565b93919750945060409150610883565b5050505093919750945060409150610883565b61095191978b0261073f565b958c6108a3565b61099b919293945061099661097c61098e61098761097c6108266108268760c81c90565b62ffffff1660020b90565b9360b01c90565b62ffffff1690565b610aca565b9190935f146109eb576109ca926109c392506109bd91018b8b0360801b61073f565b92610b0d565b6002900a90565b115b6109da578a8e5f8b8f61086e565b505050939150965060409150610883565b6109c3919350610a036109bd91610a0c940160801b90565b8b8b039061073f565b106109cc565b50600160d91b8614610869565b50505050939150965060409150610883565b9392505f610815565b50916040919650610883565b9950505050509392505050565b600160d91b03610a6457620493e090565b620186a090565b5f905b8065ffffffffffff811603610a845760081b1790565b906008019060081c610a6e565b9060205b82518111610ac357828101516001600160a01b03838116911614610abb57602001610a95565b505050600190565b5050505f90565b8190818082075f8312169105030290810160020b90620d89e7198160020b125f14610af85750620d89e71991565b91620d89e88213610b0557565b620d89e89150565b60020b8060ff1d8181011890600182167001fffcb933bd6fad37aa2d162d1a59400102600160801b189160028116610df9575b60048116610ddd575b60088116610dc1575b60108116610da5575b60208116610d89575b60408116610d6d575b60808116610d51575b6101008116610d35575b6102008116610d19575b6104008116610cfd575b6108008116610ce1575b6110008116610cc5575b6120008116610ca9575b6140008116610c8d575b6180008116610c71575b620100008116610c55575b620200008116610c3a575b620400008116610c1f575b6208000016610c06575b5f12610bfe575b60401c90565b5f1904610bf8565b6b048a170391f7dc42444e8fa290910260801c90610bf1565b6d2216e584f5fa1ea926041bedfe9890920260801c91610be7565b916e5d6af8dedb81196699c329225ee6040260801c91610bdc565b916f09aa508b5b7a84e1c677de54f3e99bc90260801c91610bd1565b916f31be135f97d08fd981231505542fcfa60260801c91610bc6565b916f70d869a156d2a1b890bb3df62baf32f70260801c91610bbc565b916fa9f746462d870fdf8a65dc1f90e061e50260801c91610bb2565b916fd097f3bdfd2022b8845ad8f792aa58250260801c91610ba8565b916fe7159475a2c29b7443b29c7fa6e889d90260801c91610b9e565b916ff3392b0822b70005940c7a398e4b70f30260801c91610b94565b916ff987a7253ac413176f2b074cf7815e540260801c91610b8a565b916ffcbe86c7900a88aedcffc83b479aa3a40260801c91610b80565b916ffe5dee046a99a2a811c461f1969c30530260801c91610b76565b916fff2ea16466c96a3843ec78b326b528610260801c91610b6d565b916fff973b41fa98c081472e6896dfb254c00260801c91610b64565b916fffcb9843d60f6159c9db58835c9266440260801c91610b5b565b916fffe5caca7e10e4e61c3624eaa0941cd00260801c91610b52565b916ffff2e50f5f656932ef12357cf3c7fdcc0260801c91610b49565b916ffff97272373d413259a46990580e213a0260801c91610b405600000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001")
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
