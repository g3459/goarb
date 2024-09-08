package caller

import (
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/g3459/goarb/contracts/interfaces"
	"github.com/g3459/goarb/utils"
)

type Route struct {
	AmOut *big.Int
	Calls []byte
}

type Protocol struct {
	Factory  *common.Address `json:"factory"`
	InitCode *common.Hash    `json:"initCode"`
}

type Protocols struct {
	UniV2     []Protocol `json:"uniV2"`
	UniV3     []Protocol `json:"uniV3"`
	AlgebraV3 []Protocol `json:"algebraV3"`
}

type Step struct {
	Element  rpc.BatchElem
	Decode   func(interface{}) interface{}
	Callback func(interface{})
}

type Batch []*Step

func S(res interface{}, decoder func(interface{}) interface{}, callback func(interface{}), method string, args ...interface{}) *Step {
	return &Step{rpc.BatchElem{Method: method, Args: args, Result: res}, decoder, callback}
}

func (batch Batch) Call(txParams map[string]interface{}, block string, decoder func(interface{}) interface{}, callback func(interface{})) Batch {
	return append(batch, S(new(string), decoder, callback, "eth_call", txParams, block))
}

func (batch Batch) BalanceOf(token *common.Address, account *common.Address, block string, callback func(interface{})) Batch {
	data, _ := interfaces.Erc20ABI.Pack("balanceOf", account)
	return batch.Call(map[string]interface{}{"to": token, "input": hexutil.Encode(data)}, block, bigIntDecoder, callback)
}

func (batch Batch) FindPools(token0 *common.Address, token1 *common.Address, protocols Protocols, poolFinder *common.Address, block string, callback func(interface{})) Batch {
	data, err := interfaces.PoolFinderABI.Pack("findPools", token0, token1, protocols)
	if err != nil {
		log.Println("zssdfsf", err)
	}
	return batch.Call(map[string]interface{}{"to": poolFinder, "input": hexutil.Encode(data)}, block, poolsDecoder, callback)
}

func (batch Batch) FindRoutes(tIn *big.Int, amIn *big.Int, pools [][][]*big.Int, gasPrice *big.Int, router *common.Address, block string, callback func(interface{})) Batch {
	data, _ := interfaces.RouterABI.Pack("findRoutes", tIn, amIn, pools)
	return batch.Call(map[string]interface{}{"to": router, "input": hexutil.Encode(data)}, block, routesDecoder, callback)
}

func (batch Batch) GasPrice(callback func(interface{})) Batch {
	return append(batch, S(new(string), bigIntDecoder, callback, "eth_gasPrice"))
}

func (batch Batch) BlockNumber(callback func(interface{})) Batch {
	return append(batch, S(new(string), uint64Decoder, callback, "eth_blockNumber"))
}

func (batch Batch) Nonce(account *common.Address, block string, callback func(interface{})) Batch {
	return append(batch, S(new(string), uint64Decoder, callback, "eth_getTransactionCount", account, block))
}

func (batch Batch) ExecutePoolCalls(calls []byte, caller *common.Address, minerTip *big.Int, maxFeePerGas *big.Int, nonce uint64, chainId *big.Int, privateKey *common.Hash, callback func(interface{})) Batch {
	return batch.SendRawTx(utils.SignTx(&types.DynamicFeeTx{ChainID: chainId, Nonce: nonce, GasTipCap: minerTip, GasFeeCap: maxFeePerGas, Gas: utils.RouteGas(calls), To: caller, Value: new(big.Int), Data: calls, AccessList: utils.AccessListForCalls(calls)}, privateKey), callback)
}

func (batch Batch) ExecuteCall(to *common.Address, call []byte, caller *common.Address, minerTip *big.Int, maxFeePerGas *big.Int, nonce uint64, chainId *big.Int, privateKey *common.Hash, callback func(interface{})) Batch {
	data, _ := interfaces.CallerABI.Pack("execute", to, call)
	return batch.SendRawTx(utils.SignTx(&types.DynamicFeeTx{ChainID: chainId, Nonce: nonce, GasTipCap: minerTip, GasFeeCap: maxFeePerGas, Gas: 1000000, To: caller, Value: new(big.Int), Data: data}, privateKey), callback)
}

func (batch Batch) Transfer(caller *common.Address, token *common.Address, to *common.Address, amount *big.Int, minerTip *big.Int, maxFeePerGas *big.Int, nonce uint64, chainId *big.Int, privateKey *common.Hash, callback func(interface{})) Batch {
	data, _ := interfaces.Erc20ABI.Pack("transfer", to, amount)
	return batch.ExecuteCall(token, data, caller, minerTip, maxFeePerGas, nonce, chainId, privateKey, callback)
}

func (batch Batch) SendRawTx(rawTx string, callback func(interface{})) Batch {
	return append(batch, S(new(string), nil, callback, "eth_sendRawTransaction", rawTx))
}

func (batch Batch) LogsByTopic(topics [][]string, fromBlock string, toBlock string, callback func(interface{})) Batch {
	return append(batch, S(new([]interface{}), nil, callback, "eth_getLogs", map[string]interface{}{"fromBlock": fromBlock, "toBlock": toBlock, "topics": topics}))
}

func (batch Batch) BlockByNumber(block string, callback func(interface{})) Batch {
	return append(batch, S(new(map[string]interface{}), nil, callback, "eth_getBlockByNumber", block, false))
}

func (batch Batch) Submit(rpcclient *rpc.Client) ([]interface{}, error) {
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
			if batch[i].Decode != nil {
				res[i] = batch[i].Decode(batchElems[i].Result)
			} else {
				res[i] = batchElems[i].Result
			}
		} else {
			res[i] = batchElems[i].Error
		}
		if batch[i].Callback != nil {
			batch[i].Callback(res[i])
		}
	}
	return res, nil
}
