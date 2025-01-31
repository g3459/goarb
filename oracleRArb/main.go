package main

import (
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
	PrivateKeys   []*common.Hash    `json:"privateKeys"`
	PoolFinder    *common.Address   `json:"poolFinder"`
	Caller        *common.Address   `json:"caller"`
	TokenConfs    []TokenConf       `json:"tokens"`
	RpcUrls       []string          `json:"rpcUrls"`
	ChainId       *big.Int          `json:"chainId"`
	MinEth        *big.Int          `json:"minEth"`
	MinLiqEth     *big.Int          `json:"minLiqEth"`
	MaxGasPrice   *big.Int          `json:"maxGasPrice"`
	GasFee        uint              `json:"gasFee"`
	ExchangeFee   float64           `json:"exchangeFee"`
	MinBen        *big.Int          `json:"minBen"`
	MinRatio      float64           `json:"minRatio"`
	Protocols     []caller.Protocol `json:"protocols"`
	FakeBalance   bool              `json:"fakeBalance"`
	LogLevel      int               `json:"logLevel"`
	RouteMaxLen   uint8             `json:"routeMaxLen"`
	Polling       time.Duration     `json:"polling"`
	ExecTime      time.Duration     `json:"execTime"`
	IsOpRollup    bool              `json:"isOpRollup"`
	MaxL1GasPrice *big.Int          `json:"maxL1GasPrice"`
	L1GasFee      uint              `json:"L1GasFee"`
	L1GasMult     float64           `json:"L1GasMult"`
	PendingBlock  bool              `json:"pendingBlock"`
	//RouteDepth  uint8             `json:"routeDepth"`
	//LogFile     string            `json:"logFile"`
	//Timeout     time.Duration     `json:"timeout"`
}

var (
	conf              Configuration
	router            *common.Address
	ethPriceX64Oracle []*big.Int
	rpcClients        = make(map[string]*rpc.Client)
	rpcClientsBanMap  = map[*rpc.Client]time.Time{}
	simClient         *rpc.Client
	hBlockn           uint64
	poolStatesBanMap  = make(map[[4]byte]bool)
	//logFile           *os.File
	routerBytecode, _ = hex.DecodeString("60a034607d57601f610c4f38819003918201601f19168301916001600160401b03831184841017608157808492602094604052833981010312607d57518060051b9080820460201490151715606957608052604051610bb990816100968239608051816104eb0152f35b634e487b7160e01b5f52601160045260245ffd5b5f80fd5b634e487b7160e01b5f52604160045260245ffdfe60806040526004361015610011575f80fd5b5f3560e01c63292b901314610024575f80fd5b346101aa5760603660031901126101aa5760043567ffffffffffffffff81116101aa57366023820112156101aa578060040135610060816101e4565b9161006e60405193846101c2565b8183526024602084019260051b820101903682116101aa5760248101925b8284106100bb576100b76100ab866024356100a56101fc565b916102bd565b6040519182918261020c565b0390f35b833567ffffffffffffffff81116101aa578201366043820112156101aa5760248101356100e7816101e4565b916100f560405193846101c2565b818352602060248185019360051b83010101903682116101aa5760448101925b82841061012f57505050908252506020938401930161008c565b833567ffffffffffffffff81116101aa5760249083010136603f820112156101aa5760208101359167ffffffffffffffff83116101ae5760405161017d601f8501601f1916602001826101c2565b83815236604084860101116101aa575f602085819660408397018386013783010152815201930192610115565b5f80fd5b634e487b7160e01b5f52604160045260245ffd5b90601f8019910116810190811067ffffffffffffffff8211176101ae57604052565b67ffffffffffffffff81116101ae5760051b60200190565b6044359060ff821682036101aa57565b602081016020825282518091526040820191602060408360051b8301019401925f915b83831061023e57505050505090565b909192939460208080600193603f19868203018752818a518051918291828552018484015e5f828201840152601f01601f19160101970195949190910192019061022f565b634e487b7160e01b5f52603260045260245ffd5b8051156102a45760200190565b610283565b80518210156102a45760209160051b010190565b91906102fb60ff8451936102d0856101e4565b946102de60405196876101c2565b8086526102ed601f19916101e4565b0136602087013716836102a9565b526103068251610527565b915f198151610100031c805b61031c5750505090565b9091925f915b835183101561051c576001831b908181161561051057189361034483826102a9565b5194851580156104dc575b6104cf575f905b85518210156104bb578185146104b2575f61037186886102a9565b515115158061049a575b1561044657506103aa600161039a84610394898b6102a9565b516102a9565b51905b84610437575f5b8a6105c4565b906103b584866102a9565b5181111561042c576103f660019392610409926103d287896102a9565b526104046103e08a8a6102a9565b5191604051938491602083019190602083019252565b03601f1981018452836101c2565b610582565b61041384876102a9565b5261041e83866102a9565b5081831b17915b0190610356565b505090600190610425565b61044086610297565b516103a4565b61045083886102a9565b5151151580610482575b15610478576103aa9061047187610394868b6102a9565b519061039d565b5090600190610425565b5061049186610394858a6102a9565b5151151561045a565b506104a983610394888a6102a9565b5151151561037b565b90600190610425565b929693905060019195505b01919490610322565b91959260019195506104c6565b506104e784846102a9565b51517f00000000000000000000000000000000000000000000000000000000000000001461034f565b919592600191506104c6565b909392915080610312565b90610531826101e4565b61053e60405191826101c2565b828152809261054f601f19916101e4565b01905f5b82811061055f57505050565b806060602080938501015201610553565b805191908290602001825e015f815290565b6105a4906103f661059e94936040519586936020850190610570565b90610570565b565b81156105b0570490565b634e487b7160e01b5f52601260045260245ffd5b92939291905f9081805b875182101561082e578188019060206040830151920151946fffffffffffffffffffffffffffffffff6106018760801c90565b961691828815610825575b5061065161064261063a61062d6106266106268960a01c90565b61ffff1690565b620f42400362ffffff1690565b62ffffff1690565b6001198b0102620f4240900490565b9283880197610662898387026105a6565b91600119830199858b11156107415760ff60d81b8816908c9082158015610818575b610766575b50506106953a91610b7b565b02928980610753575b5083900360011901958587111561074157836106c86106bd8360011b90565b8481870191026105a6565b036106d38860011b90565b12610741576106eb92919060011c80920191026105a6565b036106f68460011c90565b126107335750509061071361070d60409396610b93565b60a01b90565b60016b7fffffffff0000000000000160a01b0390911617915b01906105ce565b93915094506040915061072c565b5050505093915094506040915061072c565b61075f91948c026105a6565b928961069e565b61063a6107a98b61079a6107866107916107866106266106268660c81c90565b62ffffff1660020b90565b94859360b01c90565b818082075f8312169105030290565b92156107e857506107d46107ce6107db926107c960028a89030160801b90565b6105a6565b92610846565b6002900a90565b115b610741578b5f610689565b6107d4906108086107fc6108129460801b90565b60028a890301906105a6565b930160020b610846565b106107dd565b50600160d91b8314610684565b9692505f61060c565b9594505050935061083b57565b600160ff1b90911790565b600281900b620d89e719811215610b67575050620d89e7195b60020b8060ff1d8181011890600182167001fffcb933bd6fad37aa2d162d1a59400102600160801b189160028116610b4b575b60048116610b2f575b60088116610b13575b60108116610af7575b60208116610adb575b60408116610abf575b60808116610aa3575b6101008116610a87575b6102008116610a6b575b6104008116610a4f575b6108008116610a33575b6110008116610a17575b61200081166109fb575b61400081166109df575b61800081166109c3575b6201000081166109a7575b62020000811661098c575b620400008116610971575b6208000016610958575b5f12610950575b60401c90565b5f190461094a565b6b048a170391f7dc42444e8fa290910260801c90610943565b6d2216e584f5fa1ea926041bedfe9890920260801c91610939565b916e5d6af8dedb81196699c329225ee6040260801c9161092e565b916f09aa508b5b7a84e1c677de54f3e99bc90260801c91610923565b916f31be135f97d08fd981231505542fcfa60260801c91610918565b916f70d869a156d2a1b890bb3df62baf32f70260801c9161090e565b916fa9f746462d870fdf8a65dc1f90e061e50260801c91610904565b916fd097f3bdfd2022b8845ad8f792aa58250260801c916108fa565b916fe7159475a2c29b7443b29c7fa6e889d90260801c916108f0565b916ff3392b0822b70005940c7a398e4b70f30260801c916108e6565b916ff987a7253ac413176f2b074cf7815e540260801c916108dc565b916ffcbe86c7900a88aedcffc83b479aa3a40260801c916108d2565b916ffe5dee046a99a2a811c461f1969c30530260801c916108c8565b916fff2ea16466c96a3843ec78b326b528610260801c916108bf565b916fff973b41fa98c081472e6896dfb254c00260801c916108b6565b916fffcb9843d60f6159c9db58835c9266440260801c916108ad565b916fffe5caca7e10e4e61c3624eaa0941cd00260801c916108a4565b916ffff2e50f5f656932ef12357cf3c7fdcc0260801c9161089b565b916ffff97272373d413259a46990580e213a0260801c91610892565b620d89e8121561085f5750620d89e861085f565b600160d91b03610b8c57620493e090565b620249f090565b5f905b8065ffffffffffff811603610bac5760081b1790565b906008019060081c610b9656")
)

func main() {
	startConf()
	Log(3, conf)
	go func() {
		<-time.After(conf.ExecTime * time.Second)
		os.Exit(0)
	}()
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
				var b bool
				amounts[i], b = res.(*big.Int)
				if !b {
					err = errors.New("BalanceOf " + v.Token.Hex() + " Err: " + res.(error).Error())
					return
				}
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
			var b bool
			l1GasPrice, b = res.(*big.Int)
			if !b {
				err = errors.New("L1GasPrice Err: " + res.(error).Error())
				return
			}
		})
	}
	var minGasPrice *big.Int
	batch = batch.GasPrice(func(res interface{}) {
		var b bool
		minGasPrice, b = res.(*big.Int)
		if !b {
			err = errors.New("GasPrice Err: " + res.(error).Error())
			return
		}
	})
	nonces := make([]uint64, len(conf.PrivateKeys))
	for i, v := range conf.PrivateKeys {
		sender := crypto.PubkeyToAddress(crypto.ToECDSAUnsafe(v[:]).PublicKey)
		batch = batch.Nonce(&sender, "latest", func(res interface{}) {
			var b bool
			nonces[i], b = res.(uint64)
			if !b {
				err = errors.New("Nonce Err: " + res.(error).Error())
				return
			}
		})
	}
	var rqBlock string
	if conf.PendingBlock {
		rqBlock = "pending"
	} else {
		rqBlock = "latest"
	}
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
				var pools [][][]byte
				var number uint64
				_, err2 := batch.FindPoolsCheckBlockNumber(conf.MinLiqEth, tokens, conf.Protocols, hBlockn+1, conf.PoolFinder, rqBlock, func(res interface{}) {
					_res, b := res.([]interface{})
					if !b {
						err = errors.New("FindPools Err: " + res.(error).Error())
						return
					}
					pools = _res[0].([][][]byte)
					number = _res[1].(uint64)
				}).Submit(deadline, rpcclient)
				if err2 != nil {
					banClient(rpcclient, conf.Polling*time.Millisecond*10)
					Log(2, "BatchRPC Err: ", err2)
					cancel()
					return
				}
				if err != nil {
					banClient(rpcclient, conf.Polling*time.Millisecond*10)
					Log(2, "BatchExec Err: ", err)
					cancel()
					return
				}
				if number < hBlockn || pools == nil || len(pools) == 0 {
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
				if number > hBlockn {
					hBlockn = number
				}
				Log(3, fmt.Sprintf("\nNEW_BATCH {GasPrice: %v, Block: %v, ResTime: %v}", minGasPrice, number, time.Since(sts)))
				defer func(t time.Time) { Log(3, fmt.Sprintf("END_BATCH %v\n", time.Since(t))) }(time.Now())
				for privIx, priv := range conf.PrivateKeys {
					for t0 := range pools {
						for t1 := range pools[t0] {
							//Log(3, t0, t1, len(pools[t0][t1])/64)
							for i := len(pools[t0][t1]) - 32; i >= 32; i -= 64 {
								poolState := [4]byte(pools[t0][t1][i : i+4])
								//Log(3, poolState)
								if poolStatesBanMap[poolState] {
									pools[t0][t1] = append(pools[t0][t1][:i-32], pools[t0][t1][i+32:]...)
									Log(3, "Pool Discarted:", poolState)
								}
							}
						}
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
					var txGasLimit uint
					var checkFuncs []func() = make([]func(), 0)
					callsGasPriceLimit := new(big.Int)
					gasPrice := new(big.Int).Lsh(conf.MaxGasPrice, 2)
					Log(4, "BALANCES {")
					for i := range conf.TokenConfs {
						if amounts[i] == nil {
							continue
						}
						Log(4, fmt.Sprintf("    %v: %v", conf.TokenConfs[i].Token, amounts[i]))
					}
					Log(4, "}")
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
									res, err := new(caller.Batch).FindRoutes(tInx, amIn, pools, gasPrice, router, "pending", nil).Submit(deadline, simClient)
									if err != nil {
										Log(2, "FindRoutesRPC Err: ", err)
										return
									}
									routes, b := res[0].([][]byte)
									if !b {
										Log(2, amIn, tInx, "FindRoutesExec Err: ", res[0].(error))
										return
									}
									f := func() {
										ethIn := new(big.Int).Mul(amIn, ethPriceX64Oracle[tInx])
										ethIn.Rsh(ethIn, 64)
										for tOutx, calls := range routes {

											if ethPriceX64Oracle[tOutx] == nil {
												continue
											}
											if len(calls) == 0 {
												Log(5, tInx, tOutx, amIn, gasPrice, "len(calls)==0")
												continue
											}
											if txCalls != nil && int(len(calls)/32) > int(len(txCalls)/32) {
												Log(5, tInx, tOutx, amIn, gasPrice, "len(calls)>len(txCalls)")
												continue
											}
											amOut := big.NewInt(int64((uint64(calls[len(calls)-27]) << 40) | (uint64(calls[len(calls)-26]) << 32) | (uint64(calls[len(calls)-25]))<<24 | (uint64(calls[len(calls)-24]))<<16 | (uint64(calls[len(calls)-23]))<<8 | (uint64(calls[len(calls)-22]))<<8))
											amOut.Lsh(amOut, uint(calls[len(calls)-21]))
											// ll := 0
											// if len(pools[tInx]) > 0 {
											// 	ll += len(pools[tInx][tOutx]) / 0x40
											// }
											// if len(pools[tOutx]) > 0 {
											// 	ll += len(pools[tOutx][tInx]) / 0x40
											// }
											// fmt.Println(tInx, tOutx, amIn, len(calls)/0x20, amOut, ll)
											// continue
											ethOut := new(big.Int).Mul(amOut, ethPriceX64Oracle[tOutx])
											ethOut.Rsh(ethOut, 64)
											ben := new(big.Int).Sub(ethOut, ethIn)
											if ben.Sign() < 0 {
												Log(5, tInx, tOutx, amIn, gasPrice, fmt.Sprintf("ethIn(%vwei)-ethOut(%vwei)<0", ethIn, ethOut))
												continue
											}
											if new(big.Int).Mul(ethIn, big.NewInt(int64(conf.MinRatio*(1<<32)))).Cmp(new(big.Int).Lsh(ethOut, 32)) < 0 {
												Log(5, tInx, tOutx, amIn, gasPrice, "ratio<MinRatio")
												continue
											}
											if conf.ExchangeFee != 0 {
												ben.Mul(ben, big.NewInt(int64((1+conf.ExchangeFee)*(1<<32)))).Rsh(ben, 32)
											}
											gasUsage := uint(0)
											for cIx := 0; cIx < len(calls); cIx += 0x20 {
												if calls[cIx+4] == 2 {
													gasUsage += 300000
												} else {
													gasUsage += 150000
												}
											}
											txGas := big.NewInt(int64(gasUsage))
											gasFees := new(big.Int).Mul(txGas, gasPrice)
											if new(big.Int).Sub(ben, gasFees).Sign() > 0 {
												Log(4, tInx, tOutx, amIn, gasPrice, fmt.Sprintf("ben(%vwei)-gasFees(%vwei)>0", ben, gasFees))
												continue
											}
											if conf.IsOpRollup {
												l1BaseGas := uint(16*(len(calls)+int((amIn.BitLen()+7)/8)) + 1088)
												if uint(float64(l1BaseGas)*conf.L1GasMult) > l1BaseGas+conf.L1GasFee {
													l1BaseGas = uint(float64(l1BaseGas) * conf.L1GasMult)
												} else {
													l1BaseGas += conf.L1GasFee
												}
												l1Fees := big.NewInt(int64(l1BaseGas))
												l1Fees.Mul(l1Fees, l1GasPrice)
												ben.Sub(ben, l1Fees)
											}
											if conf.MinBen != nil {
												ben.Sub(ben, conf.MinBen)
											}
											txGas.Add(txGas, big.NewInt(int64(conf.GasFee)))
											gasPriceLimit := new(big.Int).Div(ben, txGas)
											if gasPriceLimit.Cmp(minGasPrice) < 0 {
												Log(4, tInx, tOutx, amIn, gasPrice, fmt.Sprintf("gasPriceLimit(%v)<minGasPrice(%v)", gasPriceLimit, minGasPrice))
												continue
											}
											if gasPriceLimit.Cmp(conf.MaxGasPrice) > 0 {
												Log(4, tInx, tOutx, amIn, gasPrice, fmt.Sprintf("gasPriceLimit(%v)>MaxGasPrice(%v)", gasPriceLimit, conf.MaxGasPrice))
												continue
											}
											if gasPriceLimit.Cmp(callsGasPriceLimit) < 0 {
												Log(4, tInx, tOutx, amIn, gasPrice, fmt.Sprintf("gasPriceLimit(%v)<callsGasPriceLimit(%v)", gasPriceLimit, callsGasPriceLimit))
												continue
											}
											Log(4, tInx, tOutx, amIn, gasPrice, calls, gasPriceLimit)
											callsGasPriceLimit = gasPriceLimit
											txCalls = append(calls, amIn.Bytes()...)
											txGasLimit = gasUsage + conf.GasFee
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
					for time.Now().Compare(dlt) <= 0 {
						for _, f := range checkFuncs {
							f()
						}
						if txCalls != nil {
							Log(1, txCalls, callsGasPriceLimit, number)
							if !conf.FakeBalance {
								if callsGasPriceLimit.Cmp(conf.MaxGasPrice) > 0 {
									callsGasPriceLimit.Set(conf.MaxGasPrice)
								}
								txb := new(caller.Batch).SendTx(&types.DynamicFeeTx{ChainID: conf.ChainId, Nonce: nonces[privIx], GasTipCap: callsGasPriceLimit, GasFeeCap: callsGasPriceLimit, Gas: uint64(txGasLimit), To: conf.Caller, Value: new(big.Int), Data: txCalls, AccessList: AccessListForCalls(txCalls)}, priv, nil)
								for _, rpcclient := range rpcClients {
									go func(rpcclient *rpc.Client) {
										res, err := txb.Submit(context.Background(), rpcclient)
										if err != nil {
											Log(3, "ExecutePoolCallsRPC Err: ", err)
											return
										}
										r, b := res[0].(*interface{})
										if !b {
											Log(3, "ExecutePoolCallsSend Err: ", res[0].(error))
											return
										}
										Log(3, (*r).(string), number)
									}(rpcclient)
								}
								for i := 0; i < len(txCalls)-32; i += 32 {
									poolState := [4]byte(txCalls[i : i+4])
									poolState[0] &= 0x7f
									poolStatesBanMap[poolState] = true
									Log(3, "Pool Banned:", poolState)
								}
							}
							break
						}
						<-time.After(time.Millisecond * 100)
					}
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
}

func startRpcClients(rpcUrls []string) {
	var err error
	sim := simulated.NewSimulated()
	simClient = sim.Client().Client()
	// b := caller.Batch{}
	// res, err := b.SendTx(&types.DynamicFeeTx{ChainID: big.NewInt(1337), Nonce: 0, GasTipCap: new(big.Int), GasFeeCap: new(big.Int), Gas: 10000000, Value: new(big.Int), Data: bytecodes.RouterBytecode}, conf.PrivateKey, nil).Submit(context.Background(), simClient)
	// Log(0, *res[0].(*string))
	parsedMaxLen := make([]byte, 32)
	parsedMaxLen[31] = conf.RouteMaxLen
	router, err = simulated.DeployContract(sim, append(routerBytecode, parsedMaxLen...))
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
