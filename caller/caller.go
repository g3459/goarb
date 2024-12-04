package caller

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/g3459/goarb/utils"
)

type Route struct {
	AmOut *big.Int `json:"amOut"`
	Calls []byte   `json:"calls"`
}

type Protocol struct {
	Factory *common.Address `json:"factory"`
	Id      uint8           `json:"id"`
}

type step struct {
	e *rpc.BatchElem
	d func(interface{}) interface{}
	c func(interface{})
}

type Batch []step

func S(res interface{}, decoder func(interface{}) interface{}, callback func(interface{}), method string, args ...interface{}) step {
	return step{&rpc.BatchElem{Method: method, Args: args, Result: res}, decoder, callback}
}

func (batch Batch) Call(txParams map[string]interface{}, block string, decoder func(interface{}) interface{}, callback func(interface{})) Batch {
	return append(batch, S(new(string), decoder, callback, "eth_call", txParams, block))
}

func (batch Batch) BalanceOf(token *common.Address, account *common.Address, block string, callback func(interface{})) Batch {
	data, err := Erc20ABI.Pack("balanceOf", account)
	if err != nil {
		panic(err)
	}
	return batch.Call(map[string]interface{}{"to": token, "input": hexutil.Encode(data)}, block, bigIntDecoder, callback)
}

func (batch Batch) FindPools(minEth *big.Int, tokens []common.Address, protocols []Protocol, poolFinder *common.Address, block string, callback func(interface{})) Batch {
	parsed := make([]*big.Int, len(protocols))
	for i, v := range protocols {
		parsed[i] = new(big.Int).SetBytes(append([]byte{v.Id}, v.Factory.Bytes()...))
	}
	data, _ := PoolFinderABI.Pack("findPools", minEth, tokens, parsed)
	return batch.Call(map[string]interface{}{"to": poolFinder, "input": hexutil.Encode(data)}, block, poolsDecoder, callback)
}

func (batch Batch) FindRoutes(maxLen uint8, tIn uint8, amIn *big.Int, pools [][][]byte, gasPrice *big.Int, router *common.Address, block string, callback func(interface{})) Batch {
	data, _ := RouterABI.Pack("findRoutes", maxLen, tIn, amIn, pools)
	return batch.Call(map[string]interface{}{"to": router, "gasPrice": hexutil.EncodeBig(gasPrice), "input": hexutil.Encode(data)}, block, routesDecoder, callback)
}

func (batch Batch) EthBalance(account *common.Address, block string, callback func(interface{})) Batch {
	return append(batch, S(new(string), bigIntDecoder, callback, "eth_getBalance", account, block))
}

func (batch Batch) GasPrice(callback func(interface{})) Batch {
	return append(batch, S(new(string), bigIntDecoder, callback, "eth_gasPrice"))
}

func (batch Batch) ChainId(callback func(interface{})) Batch {
	return append(batch, S(new(string), uint64Decoder, callback, "eth_chainID"))
}

func (batch Batch) BlockNumber(callback func(interface{})) Batch {
	return append(batch, S(new(string), uint64Decoder, callback, "eth_blockNumber"))
}

func (batch Batch) Nonce(account *common.Address, block string, callback func(interface{})) Batch {
	return append(batch, S(new(string), uint64Decoder, callback, "eth_getTransactionCount", account, block))
}

func (batch Batch) SendTx(tx *types.DynamicFeeTx, privateKey *common.Hash, callback func(interface{})) Batch {
	return batch.SendRawTx(utils.SignTx(tx, privateKey), callback)
}

func (batch Batch) ExecuteCall(to *common.Address, call []byte, caller *common.Address, minerTip *big.Int, maxFeePerGas *big.Int, nonce uint64, chainId *big.Int, privateKey *common.Hash, callback func(interface{})) Batch {
	data, _ := CallerABI.Pack("execute", to, call)
	return batch.SendTx(&types.DynamicFeeTx{ChainID: chainId, Nonce: nonce, GasTipCap: minerTip, GasFeeCap: maxFeePerGas, Gas: 1000000, To: caller, Value: new(big.Int), Data: data}, privateKey, callback)
}

func (batch Batch) ExecuteTransfer(caller *common.Address, token *common.Address, to *common.Address, amount *big.Int, minerTip *big.Int, maxFeePerGas *big.Int, nonce uint64, chainId *big.Int, privateKey *common.Hash, callback func(interface{})) Batch {
	data, _ := Erc20ABI.Pack("transfer", to, amount)
	return batch.ExecuteCall(token, data, caller, minerTip, maxFeePerGas, nonce, chainId, privateKey, callback)
}

func (batch Batch) ExecuteApprove(caller *common.Address, token *common.Address, spender *common.Address, amount *big.Int, minerTip *big.Int, maxFeePerGas *big.Int, nonce uint64, chainId *big.Int, privateKey *common.Hash, callback func(interface{})) Batch {
	data, _ := Erc20ABI.Pack("approve", spender, amount)
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

func (batch Batch) Submit(ctx context.Context, rpcclient *rpc.Client) ([]interface{}, error) {
	batchElems := make([]rpc.BatchElem, len(batch))
	for i := range batch {
		batchElems[i] = *batch[i].e
	}
	err := rpcclient.BatchCallContext(ctx, batchElems)
	if err != nil {
		return nil, err
	}
	res := make([]interface{}, len(batchElems))
	for i := range batchElems {
		if batchElems[i].Error == nil {
			if batch[i].d != nil {
				res[i] = batch[i].d(batchElems[i].Result)
			} else {
				res[i] = batchElems[i].Result
			}
		} else {
			res[i] = batchElems[i].Error
		}
		if batch[i].c != nil {
			batch[i].c(res[i])
		}
	}
	return res, nil
}
