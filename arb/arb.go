package main

import (
	//"context"

	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"math"
	"math/big"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/g3459/goarb/caller"
	"github.com/g3459/goarb/utils"
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
	PrivateKey   *common.Hash      `json:"privateKey"`
	PoolFinder   *common.Address   `json:"poolFinder"`
	Caller       *common.Address   `json:"caller"`
	TokenConfs   []TokenConf       `json:"tokens"`
	RpcProviders []string          `json:"rpcProviders"`
	ChainId      *big.Int          `json:"chainId"`
	MaxGasPrice  *big.Int          `json:"maxGasPrice"`
	MinEth       *big.Int          `json:"minEth"`
	MaxMinerTip  *big.Int          `json:"maxMinerTip"`
	MinMinerTip  *big.Int          `json:"minMinerTip"`
	MinGasBen    *big.Int          `json:"minGasBen"`
	Protocols    []caller.Protocol `json:"protocols"`
	FakeBalance  bool              `json:"fakeBalance"`
	LogLevel     int               `json:"logLevel"`
	RouteDepth   uint8             `json:"routeDepth"`
	RouteMaxLen  uint8             `json:"routeMaxLen"`
	LogFile      string            `json:"logFile"`
	Timeout      time.Duration     `json:"timeout"`
	ExecTimeout  time.Duration     `json:"execTimeout"`
}

var (
	conf              Configuration
	router            = common.HexToAddress("0x8988167E088c87Cd314Df6d3C2b83da5aCb93AcE")
	ethPriceX64Oracle []*big.Int
	rpcclients        = make(map[string]*rpc.Client)
	simClient         *rpc.Client
	hNumber           uint64
	sender            common.Address
	logFile           *os.File
	lastCalls         []byte
)

func main() {
	readConf()
	execTimeout()
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
	startRPCProviders()
	//batch declaration
	var err error
	batch := caller.Batch{}
	amounts := make([]*big.Int, len(conf.TokenConfs))
	if !conf.FakeBalance {
		for i, v := range conf.TokenConfs {
			if !v.Oracle.Active {
				continue
			}
			batch = batch.BalanceOf(v.Token, conf.Caller, "latest", func(res interface{}) {
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
	batch = batch.BlockByNumber("latest", func(res interface{}) {
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
	batch = batch.FindPools(new(big.Int).Lsh(conf.MinEth, 1), tokens, conf.Protocols, conf.PoolFinder, "pending", func(res interface{}) {
		var b bool
		pools, b = res.([][][]byte)
		if !b {
			err = errors.New("FindPools Err: " + res.(error).Error())
			return
		}
	})
	nonce := uint64(0)
	batch = batch.Nonce(&sender, "latest", func(res interface{}) {
		var b bool
		nonce, b = res.(uint64)
		if !b {
			err = errors.New("Nonce Err: " + res.(error).Error())
			return
		}
	})
	///
	//logic execution
	banMap := map[string]time.Time{}
	for {
		for k, rpcclient := range rpcclients {
			if time.Now().Compare(banMap[k]) < 0 {
				continue
			}
			err = nil
			deadline, cancel := context.WithDeadline(context.Background(), time.Now().Add(conf.Timeout*time.Millisecond))
			sts2 := time.Now()
			_, err2 := batch.Submit(deadline, rpcclient)
			if err2 != nil {
				banMap[k] = time.Now().Add(conf.Timeout * time.Millisecond * 60)
				Log(1, "BatchRPC Err: ", err2)
				continue
			}
			if err != nil {
				banMap[k] = time.Now().Add(conf.Timeout * time.Millisecond * 60)
				Log(0, "BatchExec Err: ", err)
				continue
			}
			sts := time.Now()
			if sts.Sub(sts2) > conf.Timeout*time.Millisecond {
				log.Println(k)
				Log(0, "BatchCall Timeout: ", sts.Sub(sts2))
				continue
			}
			if number < hNumber {
				continue
			}
			if number > hNumber {
				hNumber = number
			}
			// token := common.HexToAddress("0x8f3Cf7ad23Cd3CaDbD9735AFf958023239c6A063")
			// amount := big.NewInt(59067310702122457)
			// amount.Mul(amount, big.NewInt(1000))
			// caller.Batch{}.Transfer(conf.Caller, &token, &sender, amount, conf.MinMinerTip, conf.MaxGasPrice, nonce, conf.ChainId, conf.PrivateKey, nil).Submit(context.Background(), rpcclient)
			// continue
			Log(4, "START", k, number, sts.Sub(sts2))
			minGasPrice := new(big.Int).Add(baseFee, conf.MinMinerTip)
			var wg sync.WaitGroup
			var mu sync.Mutex
			var calls []byte
			callsGasPrice := minGasPrice
			maxItemGasPrice := minGasPrice
			for i := range conf.TokenConfs {
				if amounts[i] == nil {
					continue
				}
				Log(4, "Token:", conf.TokenConfs[i].Token, ", AmIn:", amounts[i], ", Price:", ethPriceX64Oracle[i])
				if ethPriceX64Oracle[i] == nil {
					continue
				}
				amInMin := new(big.Int).Div(new(big.Int).Lsh(conf.MinEth, 64), ethPriceX64Oracle[i])
				gasPrice := new(big.Int).Set(conf.MaxGasPrice)
				for gasPrice.Cmp(maxItemGasPrice) >= 0 {
					amIn := new(big.Int).Set(amounts[i])
					for amIn.Cmp(amInMin) > 0 {
						wg.Add(1)
						go func(amIn *big.Int, tInx uint8, gasPrice *big.Int) {
							defer wg.Done()
							res, err := new(caller.Batch).FindRoutes(2, tInx, amIn, pools, gasPrice, &router, "pending", nil).Submit(deadline, simClient)
							if err != nil {
								Log(0, "FindRoutesRPC Err: ", err)
								return
							}
							routes, b := res[0].([]caller.Route)
							if !b {
								Log(0, amIn, tInx, "FindRoutesExec Err: ", res[0].(error))
								return
							}
							mu.Lock()
							for tOutx, route := range routes {
								// log.Println(item.tInx, tOutx, route.AmOut, len(route.Calls))
								if ethPriceX64Oracle[tOutx] == nil || len(route.Calls) == 0 {
									continue
								}
								ethIn := new(big.Int).Mul(amIn, ethPriceX64Oracle[tInx])
								ethOut := new(big.Int).Mul(route.AmOut, ethPriceX64Oracle[tOutx])
								ethIn.Rsh(ethIn, 64)
								ethOut.Rsh(ethOut, 64)
								ben := new(big.Int).Sub(ethOut, ethIn)
								txGas := big.NewInt(int64(utils.CallsGas(route.Calls)))
								if new(big.Int).Sub(ben, new(big.Int).Mul(txGas, gasPrice)).Sign() > 0 {
									continue
								}
								txGas.Add(txGas, conf.MinGasBen)
								gasPriceLimit := new(big.Int).Div(ben, txGas)
								if gasPriceLimit.Cmp(minGasPrice) < 0 {
									continue
								}
								if gasPrice.Cmp(maxItemGasPrice) == 0 && gasPriceLimit.Cmp(callsGasPrice) < 0 {
									continue
								}
								maxItemGasPrice = gasPrice
								callsGasPrice = gasPriceLimit
								calls = route.Calls
							}
							mu.Unlock()
						}(amIn, uint8(i), gasPrice)
						Log(4, "START", i, amIn, gasPrice)
						amIn = new(big.Int).Rsh(amIn, 1)
					}
					wg.Wait()
					gasPrice = new(big.Int).Rsh(gasPrice, 1)
				}
			}
			ets := time.Now()
			Log(4, "END", number, ets.Sub(sts2))
			if calls != nil {
				Log(1, calls, minGasPrice, callsGasPrice)
				if !conf.FakeBalance {
					if bytes.Equal(calls, lastCalls) {
						Log(0, "Repeated Call")
					} else {
						lastCalls = calls
						minerTip := new(big.Int).Sub(callsGasPrice, baseFee)
						if minerTip.Cmp(conf.MaxMinerTip) > 0 {
							minerTip = conf.MaxMinerTip
						}
						if callsGasPrice.Cmp(conf.MaxGasPrice) > 0 {
							callsGasPrice = conf.MaxGasPrice
						}
						b := new(caller.Batch).ExecutePoolCalls(calls, conf.Caller, minerTip, callsGasPrice, nonce, conf.ChainId, conf.PrivateKey, nil)
						for _, rpcclient := range rpcclients {
							go func(_rpcclient *rpc.Client) {
								res, err := b.Submit(context.Background(), _rpcclient)
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
			}
			<-deadline.Done()
			cancel()
		}
	}
	///
}

func readConf() {
	rawConf, err := os.ReadFile(os.Args[1])
	if err != nil {
		Log(-1, "ReadConfFile Err: ", err)
	}
	json.Unmarshal(rawConf, &conf)
	sender = crypto.PubkeyToAddress(crypto.ToECDSAUnsafe((conf.PrivateKey)[:]).PublicKey)
}

func execTimeout() {
	go func() {
		<-time.After(conf.ExecTimeout * time.Second)
		Log(0, "exit")
		os.Exit(0)
	}()
}

func startRPCProviders() {
	var err error
	simClient, err = rpc.Dial("ws://localhost:8546")
	if err != nil {
		Log(-1, "simrpcDial Err: ", err)
	}
	for _, url := range conf.RpcProviders {
		deadline, cancel := context.WithDeadline(context.Background(), time.Now().Add(1000*time.Millisecond))
		client, err := rpc.DialContext(deadline, url)
		cancel()
		if err != nil {
			Log(2, "rpcDial Err: ", err, url)
			continue
		}
		Log(2, "rpcDial: ", url)
		rpcclients[url] = client
	}
}

func startUsdOracles() {
	ethPriceX64Oracle = make([]*big.Int, len(conf.TokenConfs))
	ethPriceX64Oracle[0] = new(big.Int).Lsh(big.NewInt(1), 64)
	for _, v := range conf.TokenConfs {
		if v.Oracle.Active && len(v.Oracle.Name) > 0 && v.Oracle.Name != "usd" {
			err := startBinanceUsdOracle(v.Oracle.Name)
			if err != nil {
				Log(-1, "binanceDial Err: ", err)
			} else {
				Log(2, "binanceDial: ", v.Oracle.Name)
				continue
			}
			err = startBybitUsdOracle(v.Oracle.Name)
			if err != nil {
				Log(-1, "bybitDial Err: ", err)
			} else {
				Log(2, "bybitDial: ", v.Oracle.Name)
				continue
			}

		}
	}
}

func startBinanceUsdOracle(baseToken string) error {
	wsURL := "wss://stream.binance.com:9443/ws/" + strings.ToLower(baseToken) + "usdt@aggTrade"
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return err
	}
	go func() {
		defer c.Close()
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
	wsURL := "wss://stream.bybit.com/v5/public/spot"
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return err
	}
	go func() {
		defer c.Close()
		err = c.WriteJSON(struct {
			Op   string   `json:"op"`
			Args []string `json:"args"`
		}{"subscribe", []string{"tickers." + strings.ToUpper(baseToken) + "USDT"}})
		if err != nil {
			Log(-1, "bybitWrite Err:", err)
		}
		go func() {
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
						Log(3, "bybitMsg:", "tickers."+strings.ToUpper(baseToken)+"USDT "+res["ret_msg"].(string))
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
	return nil
}

var usdEthPrice float64

func updatePriceUsdOracle(baseToken string, price float64) {
	if baseToken == conf.TokenConfs[0].Oracle.Name {
		usdEthPrice = 1 / price
		ethPX64 := utils.ToX64Int(usdEthPrice)
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
			ethPX64 := utils.ToX64Int(price * usdEthPrice)
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

type any = interface{}

func Log(level int, params ...any) {
	if conf.LogLevel >= level || conf.LogLevel == 0 {
		if level < 0 {
			log.Panicln(params...)
		} else {
			log.Println(params...)
		}
	}
}
