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
	TxFee         *big.Int          `json:"txFee"`
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
	routerBytecode, _ = hex.DecodeString("60a034607a57601f6112ea38819003918201601f19168301916001600160401b03831184841017607e57808492602094604052833981010312607a57518015607a575f1981019081116066576080526040516112579081610093823960805181610a620152f35b634e487b7160e01b5f52601160045260245ffd5b5f80fd5b634e487b7160e01b5f52604160045260245ffdfe6101c06040526004361015610012575f80fd5b5f3560e01c63292b901314610025575f80fd5b3461096e57606036600319011261096e5760043560e05267ffffffffffffffff60e0511161096e5736602360e05101121561096e5767ffffffffffffffff60e051600401351161096e5736602460e0516004013560051b60e05101011161096e5760443560ff811680910361096e576100a360e05160040135610abf565b610160526100b76024359161016051610b07565b526100c760e05160040135610abf565b6101a0526100da60e05160040135610abf565b610120525f1960e05160040135610100031c60c0525b60c051610141576040518060208101602082526101205151809152604082019060206101205101905f5b818110610128575050500390f35b825184528594506020938401939092019160010161011a565b5f610180525b60e051600401356101805110156100f0576001610180511b8060c0511615610a8d5760c0511860c0526101806101805161016051610b07565b516101005261010051158015610a4b575b610a46575f610140525b60e0516004013561014051106101bc575b6001610180510161018052610147565b610140516101805114610a41575f6080525f6101e76101805160e05160040135602460e05101610b2f565b9050151580610a1b575b156109985750600161021f6102156101805160e05160040135602460e05101610b2f565b6101405191610b74565b608052905b6102346101405161016051610b07565b5190610246610140516101a051610b07565b515f60a081905290915b60805160a05110610327575060ff16925090821561031f576102786101405161016051610b07565b52610289610140516101a051610b07565b525f905b601082106102b7575b50506001610140511b60c0511760c0525b600161014051016101405261019b565b6102c76101805161012051610b07565b519161ffff8160041b93841c1661031557506101405160081b61018051600c1b1717901b6102fb6101805161012051610b07565b511761030d6101805161012051610b07565b525f80610296565b600101915061028d565b5050506102a7565b9265ffffffffffff60a05186013560301c169265ffffffffffff60a051870135169284158015610990575b610549576103738565ffffffffffff60601b60a0518a013560601b16610bb6565b946103b487620f424062ffffff61ffff60a0518d013560b01c16820316600119610100510102046001600160801b038885021660018060a01b038a16610bd4565b61ffff60a0518a013560c81c1660020b61061d575b87156105f7576103ef908287026001600160801b0316906001600160a01b03891661101a565b955b60a05189013560ff1c600a036105ee57620493e05b62ffffff61041a610180516101a051610b07565b51911601953a8702913a860293610140516105ac575b838a03948803851115610577576104818b620f424062ffffff8f61ffff9060a051013560b01c168203166001196101005101020460011b6001600160801b038685021660018060a01b038616610bd4565b8b15610587576104a7908285026001600160801b0316906001600160a01b03851661101a565b848660011b910313610577576104f78b620f424062ffffff8f61ffff9060a051013560b01c168203166001196101005101020460011c6001600160801b038685021660018060a01b038616610bd4565b908b156105545761051b93026001600160801b0316916001600160a01b031661101a565b915b60011c910313610549575050509160a05184013560ff1c91935b602060a0510160a05293929190610250565b935093915093610537565b61057193026001600160801b0316916001600160a01b0316610fb0565b9161051d565b5050505050935093915093610537565b6105a7908285026001600160801b0316906001600160a01b038516610fb0565b6104a7565b93926105ce6105e8916105c56101805161016051610b07565b51908c02610bb6565b936105df6101805161016051610b07565b51908b02610bb6565b93610430565b620249f0610406565b610617908287026001600160801b0316906001600160a01b038916610fb0565b956103f1565b73fffd8963efd1fc6a506488495d951d51639616826001600160a01b038881166401000276a21901161161097257602087901b640100000000600160c01b031680801561096e5760ff826001600160801b031060071b83811c67ffffffffffffffff1060061b1783811c63ffffffff1060051b1783811c61ffff1060041b1783811c821060031b177f07060605060205000602030205040001060502050303040105050304000000006f8421084210842108cc6318c6db6d54be85831c1c601f161a17169160808310155f146109625750607e1982011c5b800280607f1c8160ff1c1c800280607f1c8160ff1c1c800280607f1c8160ff1c1c800280607f1c8160ff1c1c800280607f1c8160ff1c1c800280607f1c8160ff1c1c80029081607f1c8260ff1c1c80029283607f1c8460ff1c1c80029485607f1c8660ff1c1c80029687607f1c8860ff1c1c80029889607f1c8a60ff1c1c80029a8b607f1c8c60ff1c1c80029c8d80607f1c9060ff1c1c800260cd1c6604000000000000169d60cc1c6608000000000000169c60cb1c6610000000000000169b60ca1c6620000000000000169a60c91c6640000000000000169960c81c6680000000000000169860c71c670100000000000000169760c61c670200000000000000169660c51c670400000000000000169560c41c670800000000000000169460c31c671000000000000000169360c21c672000000000000000169260c11c674000000000000000169160c01c6780000000000000001690607f190160401b1717171717171717171717171717693627a301d71055774c85026f028f6481ab7f045a5af012a19d003aa919810160801d60020b906fdb2df09e81959a81455e260799a0632f0160801d60020b8082145f1461093957505b60a05161ffff908b013560c81c1660020b8082075f8312169181900591909103028815610903576001600160a01b03906108e490610c83565b166001600160a01b038216105b156103c9575050935093915093610537565b61092760018060a01b039161ffff60a0518d013560c81c1660020b0160020b610c83565b166001600160a01b03821610156108f1565b906001600160a01b038981169061094f84610c83565b161161095b57506108ab565b90506108ab565b905081607f031b6106f5565b5f80fd5b6318521d4960e21b5f9081526001600160a01b038816600452602490fd5b508315610352565b6109b16101405160e05160040135602460e05101610b2f565b90501515806109f5575b156109ef576109e66109dc6101405160e05160040135602460e05101610b2f565b6101805191610b74565b60805290610224565b506102a7565b50610a126109dc6101405160e05160040135602460e05101610b2f565b905015156109bb565b50610a386102156101805160e05160040135602460e05101610b2f565b905015156101f1565b6102a7565b6101ac565b5061ffff610a5f6101805161012051610b07565b517f000000000000000000000000000000000000000000000000000000000000000060041b1c161515610191565b506101ac565b67ffffffffffffffff8111610aab5760051b60200190565b634e487b7160e01b5f52604160045260245ffd5b90610ac982610a93565b60405190601f01601f1916810167ffffffffffffffff811182821017610aab576040528281528092610afd601f1991610a93565b0190602036910137565b8051821015610b1b5760209160051b010190565b634e487b7160e01b5f52603260045260245ffd5b9190811015610b1b5760051b81013590601e198136030182121561096e57019081359167ffffffffffffffff831161096e576020018260051b3603811361096e579190565b9190811015610b1b5760051b81013590601e198136030182121561096e57019081359167ffffffffffffffff831161096e57602001823603811361096e579190565b8115610bc0570490565b634e487b7160e01b5f52601260045260245ffd5b926001600160a01b038416156001600160801b0383161517610c765715610c0157610bfe92611067565b90565b610c3a92916001600160a01b038111610c5d576001600160801b03610c2a92169060601b610bb6565b905b6001600160a01b0316611046565b6001600160a01b038116908103610c4e5790565b6393dafdf160e01b5f5260045ffd5b6001600160801b03610c7092169061114d565b90610c2c565b634f2461b85f526004601cfd5b60020b908160ff1d82810118620d89e88111610f9d5763ffffffff9192600182167001fffcb933bd6fad37aa2d162d1a59400102600160801b189160028116610f81575b60048116610f65575b60088116610f49575b60108116610f2d575b60208116610f11575b60408116610ef5575b60808116610ed9575b6101008116610ebd575b6102008116610ea1575b6104008116610e85575b6108008116610e69575b6110008116610e4d575b6120008116610e31575b6140008116610e15575b6180008116610df9575b620100008116610ddd575b620200008116610dc2575b620400008116610da7575b6208000016610d8e575b5f12610d86575b0160201c90565b5f1904610d7f565b6b048a170391f7dc42444e8fa290910260801c90610d78565b6d2216e584f5fa1ea926041bedfe9890920260801c91610d6e565b916e5d6af8dedb81196699c329225ee6040260801c91610d63565b916f09aa508b5b7a84e1c677de54f3e99bc90260801c91610d58565b916f31be135f97d08fd981231505542fcfa60260801c91610d4d565b916f70d869a156d2a1b890bb3df62baf32f70260801c91610d43565b916fa9f746462d870fdf8a65dc1f90e061e50260801c91610d39565b916fd097f3bdfd2022b8845ad8f792aa58250260801c91610d2f565b916fe7159475a2c29b7443b29c7fa6e889d90260801c91610d25565b916ff3392b0822b70005940c7a398e4b70f30260801c91610d1b565b916ff987a7253ac413176f2b074cf7815e540260801c91610d11565b916ffcbe86c7900a88aedcffc83b479aa3a40260801c91610d07565b916ffe5dee046a99a2a811c461f1969c30530260801c91610cfd565b916fff2ea16466c96a3843ec78b326b528610260801c91610cf4565b916fff973b41fa98c081472e6896dfb254c00260801c91610ceb565b916fffcb9843d60f6159c9db58835c9266440260801c91610ce2565b916fffe5caca7e10e4e61c3624eaa0941cd00260801c91610cd9565b916ffff2e50f5f656932ef12357cf3c7fdcc0260801c91610cd0565b916ffff97272373d413259a46990580e213a0260801c91610cc7565b826345c3193d60e11b5f5260045260245ffd5b906001600160a01b0380821690831611611014575b6001600160a01b03821691821561100857610bfe93611003926001600160a01b0380821693909103169060601b600160601b600160e01b03166111d7565b610bb6565b62bfc9215f526004601cfd5b90610fc5565b610bfe926001600160a01b03928316919092160360ff81901d90810118906001600160801b0316611104565b9190820180921161105357565b634e487b7160e01b5f52601160045260245ffd5b919081156110ff5760601b600160601b600160e01b0316916001600160a01b0316818102816110968483610bb6565b146110c4575b50906110ab6110b09284610bb6565b611046565b80820491061515016001600160a01b031690565b830183811061109c5791506110da8282856111d7565b928215610bc057096110f3575b6001600160a01b031690565b600101806110e7575f80fd5b505090565b81810291905f1982820991838084109303928084039384600160601b111561096e571461114457600160601b910990828211900360a01b910360601c1790565b50505060601c90565b90606082901b905f19600160601b84099282808510940393808503948584111561096e57146111d0578190600160601b900981805f03168092046002816003021880820260020302808202600203028082026002030280820260020302808202600203028091026002030293600183805f03040190848311900302920304170290565b5091500490565b91818302915f198185099383808610950394808603958685111561096e571461124f579082910981805f03168092046002816003021880820260020302808202600203028082026002030280820260020302808202600203028091026002030293600183805f03040190848311900302920304170290565b50509150049056")
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
											if conf.TxFee != nil {
												ben.Sub(ben, conf.TxFee)
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
