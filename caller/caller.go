package caller

import (
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/g3459/goarb/utils"
)

type Route struct {
	AmOut *big.Int
	Calls []byte
}

type TokenInfo struct {
	Token   common.Address `json:"token"`
	EthPX64 *big.Int       `json:"ethPX64"`
}

type Step struct {
	Element rpc.BatchElem
	Decode  func(interface{}) interface{}
}

type Batch []*Step

const ADDRZERO = "0x0000000000000000000000000000000000000000"

var erc20ABIReader, _ = os.Open("../contracts/interfaces/erc20ABI.json")
var erc20ABI, _ = abi.JSON(erc20ABIReader)

var routerABIReader, _ = os.Open("../contracts/interfaces/routerABI.json")
var routerABI, _ = abi.JSON(routerABIReader)

var callerABIReader, _ = os.Open("../contracts/interfaces/callerABI.json")
var callerABI, _ = abi.JSON(callerABIReader)

func S(res interface{}, decoder func(interface{}) interface{}, method string, args ...interface{}) *Step {
	return &Step{rpc.BatchElem{method, args, res, nil}, decoder}
}

func (batch Batch) AddCall(txParams map[string]interface{}, block string, decode func(interface{}) interface{}) Batch {
	return append(batch, S(new(string), decode, "eth_call", txParams, block))
}

func (batch Batch) AddBalances(tokens []common.Address, account common.Address) Batch {
	for i := range tokens {
		data, _ := erc20ABI.Pack("balanceOf", account)
		batch = batch.AddCall(map[string]interface{}{"from": ADDRZERO, "to": tokens[i], "input": hexutil.Encode(data)}, "latest", bigIntDecoder)
	}
	return batch
}

func (batch Batch) AddFindRoutesForAllTokensWithBalances(tokens []TokenInfo, minEth *big.Int, caller common.Address, router common.Address, gasPrice *big.Int, block string) Batch {
	data, _ := routerABI.Pack("allTokensWithBalances", tokens, minEth, caller)
	return batch.AddCall(map[string]interface{}{"to": router, "input": hexutil.Encode(data), "gasPrice": hexutil.EncodeBig(gasPrice)}, block, allTokensDecoder)
}

func (batch Batch) AddFindRoutesForSingleToken(tokens []TokenInfo, amIn *big.Int, tIn *big.Int, router common.Address, block string) Batch {
	data, _ := routerABI.Pack("singleToken", tokens, amIn, tIn)
	return batch.AddCall(map[string]interface{}{"to": router, "input": hexutil.Encode(data)}, block, singleTokenDecoder)
}

func (batch Batch) AddCallFindPools(tokens []common.Address, ethPricesX64 []*big.Int, minEth *big.Int, router common.Address, block string) Batch {
	data, _ := routerABI.Pack("findPools", tokens, ethPricesX64, minEth)
	return batch.AddCall(map[string]interface{}{"from": ADDRZERO, "to": router, "input": hexutil.Encode(data)}, block, stringDecoder)
}

func (batch Batch) AddGasPrice() Batch {
	return append(batch, S(new(string), bigIntDecoder, "eth_gasPrice"))
}

func (batch Batch) AddBlockNumber() Batch {
	return append(batch, S(new(string), uint64Decoder, "eth_blockNumber"))
}

func (batch Batch) AddNonce(account common.Address, block string) Batch {
	return append(batch, S(new(string), uint64Decoder, "eth_getTransactionCount", account, block))
}

func (batch Batch) AddExecuteRoute(calls []byte, nonce uint64, caller common.Address, minerTip *big.Int, maxFeePerGas *big.Int, chainId *big.Int, privateKey common.Hash) Batch {
	return batch.AddSendRawTx(utils.SignTx(&types.DynamicFeeTx{ChainID: chainId, Nonce: nonce, GasTipCap: minerTip, GasFeeCap: maxFeePerGas, Gas: utils.RouteGas(calls), To: &caller, Value: new(big.Int), Data: calls, AccessList: utils.AccessListForCalls(calls)}, privateKey))
}

func (batch Batch) AddExecuteCall(to common.Address, call []byte, caller common.Address, minerTip *big.Int, maxFeePerGas *big.Int, nonce uint64, chainId *big.Int, privateKey common.Hash) Batch {
	data, _ := callerABI.Pack("execute", to, call)
	return batch.AddSendRawTx(utils.SignTx(&types.DynamicFeeTx{ChainID: chainId, Nonce: nonce, GasTipCap: minerTip, GasFeeCap: maxFeePerGas, Gas: 1000000, To: &caller, Value: new(big.Int), Data: data}, privateKey))
}

func (batch Batch) AddExecuteTransfer(caller common.Address, token common.Address, to common.Address, amount *big.Int, minerTip *big.Int, gasPrice *big.Int, nonce uint64, chainId *big.Int, privateKey common.Hash) Batch {
	data, _ := erc20ABI.Pack("transfer", to, amount)
	return batch.AddExecuteCall(token, data, caller, minerTip, gasPrice, nonce, chainId, privateKey)
}

func (batch Batch) AddSendRawTx(rawTx string) Batch {
	return append(batch, S(new(string), stringDecoder, "eth_sendRawTransaction", rawTx))
}

func (batch Batch) AddLogsByTopic(topics [][]string, fromBlock string, toBlock string) Batch {
	return append(batch, S(new([]interface{}), sliceDecoder, "eth_getLogs", map[string]interface{}{"fromBlock": fromBlock, "toBlock": toBlock, "topics": topics}))
}

func (batch Batch) AddBlockByNumber(block string) Batch {
	return append(batch, S(new(map[string]interface{}), mapStringDecoder, "eth_getBlockByNumber", block, false))
}

func (batch Batch) Execute(rpcclient *rpc.Client) ([]interface{}, error) {
	batchElems := make([]rpc.BatchElem, len(batch))
	for i := range batch {
		batchElems[i] = batch[i].Element
	}
	err := rpcclient.BatchCall(batchElems)
	if err != nil {
		return nil, err
	}
	res := make([]interface{}, len(batchElems))
	for i := range batchElems {
		if batchElems[i].Error == nil {
			res[i] = batch[i].Decode(batchElems[i].Result)
		} else {
			//log.Println("Error:", batchElems[i].Error, res[i])
		}
	}
	return res, nil
}
