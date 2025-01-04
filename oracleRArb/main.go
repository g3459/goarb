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
	MinGasBen   uint64            `json:"minGasBen"`
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
	routerBytecode, _ = hex.DecodeString("60c03461008757601f610f7c38819003918201601f19168301916001600160401b0383118484101761008b57808492604094855283398101031261008757610052602061004b8361009f565b920161009f565b60a052608052604051610ecf90816100ad82396080518161096e015260a051818181610417015281816104ef01526108590152f35b5f80fd5b634e487b7160e01b5f52604160045260245ffd5b519081151582036100875756fe60806040526004361015610011575f80fd5b5f3560e01c633818b99d14610024575f80fd5b346101b85760803660031901126101b85761003d6101c1565b6100456101d1565b906064359167ffffffffffffffff83116101b857366023840112156101b857826004013561007281610217565b9361008060405195866101f5565b8185526024602086019260051b820101903682116101b85760248101925b8284106100c9576100c56100b688604435888a610398565b6040939193519384938461026c565b0390f35b833567ffffffffffffffff81116101b8578201366043820112156101b85760248101356100f581610217565b9161010360405193846101f5565b818352602060248185019360051b83010101903682116101b85760448101925b82841061013d57505050908252506020938401930161009e565b833567ffffffffffffffff81116101b85760249083010136603f820112156101b85760208101359167ffffffffffffffff83116101bc5760405161018b601f8501601f1916602001826101f5565b83815236604084860101116101b8575f602085819660408397018386013783010152815201930192610123565b5f80fd5b6101e1565b6004359060ff821682036101b857565b6024359060ff821682036101b857565b634e487b7160e01b5f52604160045260245ffd5b90601f8019910116810190811067ffffffffffffffff8211176101bc57604052565b67ffffffffffffffff81116101bc5760051b60200190565b90602080835192838152019201905f5b81811061024c5750505090565b825167ffffffffffffffff1684526020938401939092019160010161023f565b6060808252825190820181905260808201959492602001905f5b818110610316575050508085036020820152825180865260208601906020808260051b8901019501915f905b8282106102d357505050506102d0939450604081840391015261022f565b90565b9091929560208080600193601f198d8203018652818b518051918291828552018484015e5f828201840152601f01601f19160101980194939190910191016102b2565b8251885260209788019790920191600101610286565b634e487b7160e01b5f52603260045260245ffd5b80511561034d5760200190565b61032c565b805182101561034d5760209160051b010190565b9061037082610217565b61037d60405191826101f5565b828152809261038e601f1991610217565b0190602036910137565b929390936060928151906103ab82610217565b916103b960405193846101f5565b8083526103c8601f1991610217565b013660208401376103dd60ff83981683610352565b528151916103ea83610217565b926103f860405194856101f5565b808452610407601f1991610217565b015f5b81811061046f57505082957f0000000000000000000000000000000000000000000000000000000000000000610450575b9161044e939160ff879460051b166104df565b565b9160ff95508161046461044e959351610366565b96509193509161043b565b808760208093880101520161040a565b8115610489570490565b634e487b7160e01b5f52601260045260245ffd5b805191908290602001825e015f815290565b61044e906104d16104cb9493604051958693602085019061049d565b9061049d565b03601f1981018452836101f5565b9193925f198251610100031c93847f0000000000000000000000000000000000000000000000000000000000000000955b61051d5750505050505050565b91929394955f925b8551841015610840576001841b908181161561083457186105468483610352565b5115801561081d575b610812575f905b8651821015610801578185146107f8575f6105718689610352565b51511515806107e0575b15610790575060018461059884610592898c610352565b51610352565b515b8461077c576105c8895f925b85846105c18d6001196105b9828f610352565b510195610352565b519361084d565b9290918d6105d6888a610352565b5184111561076d57610681575b5050600193926106549261061161060b6104d194881901806106058b8d610352565b52610b25565b60a01b90565b90866b7fffffffff0000000000000160a01b03161790610677575b61064f6106398a8c610352565b5191604051938491602083019190602083019252565b6104af565b61065e8489610352565b526106698388610352565b5081831b17915b0190610556565b8460ff1b1761062c565b61069c61068e8b84610352565b5167ffffffffffffffff1690565b67ffffffffffffffff6106b460ff60d81b8716610b0d565b911601903a8202903a6106da6106cd61068e8c88610352565b67ffffffffffffffff1690565b02908061074f575b506106ed898b610352565b51039084031315610741579261061161060b6104d19461073260019998956107238c67ffffffffffffffff6106549b1692610352565b9067ffffffffffffffff169052565b945050509288919495506105e3565b505050505090600190610670565b61075f816107679394880261047f565b92860261047f565b5f6106e2565b50505050505090600190610670565b6105c88961078988610340565b51926105a6565b61079a8389610352565b51511515806107c8575b156107be57846107b887610592868c610352565b5161059a565b5090600190610670565b506107d786610592858b610352565b515115156107a4565b506107ef83610592888b610352565b5151151561057b565b90600190610670565b919893600191505b01929790610525565b909792600190610809565b506108288486610352565b515160ff88161461054f565b91989360019150610809565b9096959493925080610510565b949392945f5f915f905f7f0000000000000000000000000000000000000000000000000000000000000000935b8a51821015610b0057818b01604081015196906108a06001600160a01b0389168c610b4b565b610af45760200151916fffffffffffffffffffffffffffffffff6108c48460801c90565b931691828b15610aeb575b506108e76108e06108e08b60a01c90565b61ffff1690565b620f4240038087028c610905620f424088029287818501910261047f565b9685881115610ad95760ff60d81b8d16918c918b8f85158015610acc575b610a12575b50505050610945575b50505050505060409095915b01909461087a565b610953909994959699610b0d565b3a02958c806109ff575b50868a03928587038413156109ec577f00000000000000000000000000000000000000000000000000000000000000006109a8575b5050505050505060409093905f80808080610931565b6109cf9288926109c7926109bc8d60011b90565b02809201910261047f565b039160011b90565b126109dd5780808080610992565b9391975094506040915061093d565b505050509391975094506040915061093d565b610a0b91978b0261047f565b958c61095d565b610a559192939450610a50610a36610a48610a41610a366108e06108e08760c81c90565b62ffffff1660020b90565b9360b01c90565b62ffffff1690565b610b84565b9190935f14610aa557610a8492610a7d9250610a7791018b8b0360801b61047f565b92610bc7565b6002900a90565b115b610a94578a8e5f8b8f610928565b50505093915096506040915061093d565b610a7d919350610abd610a7791610ac6940160801b90565b8b8b039061047f565b10610a86565b50600160d91b8614610923565b5050505093915096506040915061093d565b9392505f6108cf565b5091604091965061093d565b9950505050509392505050565b600160d91b03610b1e57620493e090565b620186a090565b5f905b8065ffffffffffff811603610b3e5760081b1790565b906008019060081c610b28565b9060205b82518111610b7d57828101516001600160a01b03838116911614610b7557602001610b4f565b505050600190565b5050505f90565b8190818082075f8312169105030290810160020b90620d89e7198160020b125f14610bb25750620d89e71991565b91620d89e88213610bbf57565b620d89e89150565b60020b8060ff1d8181011890600182167001fffcb933bd6fad37aa2d162d1a59400102600160801b189160028116610eb3575b60048116610e97575b60088116610e7b575b60108116610e5f575b60208116610e43575b60408116610e27575b60808116610e0b575b6101008116610def575b6102008116610dd3575b6104008116610db7575b6108008116610d9b575b6110008116610d7f575b6120008116610d63575b6140008116610d47575b6180008116610d2b575b620100008116610d0f575b620200008116610cf4575b620400008116610cd9575b6208000016610cc0575b5f12610cb8575b60401c90565b5f1904610cb2565b6b048a170391f7dc42444e8fa290910260801c90610cab565b6d2216e584f5fa1ea926041bedfe9890920260801c91610ca1565b916e5d6af8dedb81196699c329225ee6040260801c91610c96565b916f09aa508b5b7a84e1c677de54f3e99bc90260801c91610c8b565b916f31be135f97d08fd981231505542fcfa60260801c91610c80565b916f70d869a156d2a1b890bb3df62baf32f70260801c91610c76565b916fa9f746462d870fdf8a65dc1f90e061e50260801c91610c6c565b916fd097f3bdfd2022b8845ad8f792aa58250260801c91610c62565b916fe7159475a2c29b7443b29c7fa6e889d90260801c91610c58565b916ff3392b0822b70005940c7a398e4b70f30260801c91610c4e565b916ff987a7253ac413176f2b074cf7815e540260801c91610c44565b916ffcbe86c7900a88aedcffc83b479aa3a40260801c91610c3a565b916ffe5dee046a99a2a811c461f1969c30530260801c91610c30565b916fff2ea16466c96a3843ec78b326b528610260801c91610c27565b916fff973b41fa98c081472e6896dfb254c00260801c91610c1e565b916fffcb9843d60f6159c9db58835c9266440260801c91610c15565b916fffe5caca7e10e4e61c3624eaa0941cd00260801c91610c0c565b916ffff2e50f5f656932ef12357cf3c7fdcc0260801c91610c03565b916ffff97272373d413259a46990580e213a0260801c91610bfa5600000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001")
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
				_, err2 := batch.FindPoolsCheckBlockNumber(conf.MinLiqEth, tokens, conf.Protocols, hBlockn+1, conf.PoolFinder, "latest", func(res interface{}) {
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
								res, err := new(caller.Batch).FindRoutes(conf.RouteMaxLen, tInx, amIn, pools, gasPrice, router, "pending", nil).Submit(deadline, simClient)
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
										txGas := big.NewInt(int64(route.GasUsage + conf.MinGasBen))
										gasFees := new(big.Int).Mul(txGas, gasPrice)
										if new(big.Int).Sub(ben, gasFees).Sign() > 0 {
											Log(4, tInx, tOutx, amIn, gasPrice, fmt.Sprintf("ben(%vwei)-gasFees(%vwei)>0", ben, gasFees))
											continue
										}
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
										txGasLimit = route.GasUsage + 150000
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
									hash, b := (*res[0].(*interface{})).(*string)
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
