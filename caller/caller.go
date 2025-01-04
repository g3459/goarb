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
	AmOut    *big.Int `json:"amOut"`
	Calls    []byte   `json:"calls"`
	GasUsage uint64   `json:"gasUsage"`
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

var (
	opGPOAddr = common.HexToAddress("0x420000000000000000000000000000000000000F")
)

func encodeProtocols(protocols []Protocol) []*big.Int {
	parsed := make([]*big.Int, len(protocols))
	for i, v := range protocols {
		t := [12]byte{}
		t[4] = v.Id
		parsed[i] = new(big.Int).SetBytes(append(t[:], v.Factory.Bytes()...))
	}
	return parsed
}

func S(decoder func(interface{}) interface{}, callback func(interface{}), method string, args ...interface{}) step {
	return step{&rpc.BatchElem{Method: method, Args: args, Result: new(interface{})}, decoder, callback}
}

func (batch Batch) Call(txParams map[string]interface{}, block string, decoder func(interface{}) interface{}, callback func(interface{})) Batch {
	return append(batch, S(decoder, callback, "eth_call", txParams, block))
}

func (batch Batch) BalanceOf(token *common.Address, account *common.Address, block string, callback func(interface{})) Batch {
	data, err := Erc20ABI.Pack("balanceOf", account)
	if err != nil {
		panic(err)
	}
	return batch.Call(map[string]interface{}{"to": token, "input": hexutil.Encode(data)}, block, bigIntDecoder, callback)
}

func (batch Batch) FindPools(minLiqEth *big.Int, tokens []common.Address, protocols []Protocol, poolFinder *common.Address, block string, callback func(interface{})) Batch {
	data, err := PoolFinderABI.Pack("findPools", minLiqEth, tokens, encodeProtocols(protocols))
	if err != nil {
		panic(err)
	}
	return batch.Call(map[string]interface{}{"to": poolFinder, "input": hexutil.Encode(data)}, block, findPoolsDecoder, callback)
}

func (batch Batch) FindPoolsCheckBlockNumber(minLiqEth *big.Int, tokens []common.Address, protocols []Protocol, minBlockNumber uint64, poolFinder *common.Address, block string, callback func(interface{})) Batch {
	data, err := PoolFinderABI.Pack("findPoolsCheckBlockNumber", minLiqEth, tokens, encodeProtocols(protocols), minBlockNumber)
	if err != nil {
		panic(err)
	}
	return batch.Call(map[string]interface{}{"to": poolFinder, "input": hexutil.Encode(data)}, block, findPoolsCheckBlockNumberDecoder, callback)
}

func (batch Batch) FindRoutes(maxLen uint8, tIn uint8, amIn *big.Int, pools [][][]byte, gasPrice *big.Int, router *common.Address, block string, callback func(interface{})) Batch {
	data, err := RouterABI.Pack("findRoutes", maxLen, tIn, amIn, pools)
	if err != nil {
		panic(err)
	}
	return batch.Call(map[string]interface{}{"to": router, "gasPrice": hexutil.EncodeBig(gasPrice), "input": hexutil.Encode(data)}, block, findRoutesDecoder, callback)
}

func (batch Batch) L1GasPrice(callback func(interface{})) Batch {
	data, err := OpGPOABI.Pack("l1BaseFee")
	if err != nil {
		panic(err)
	}
	return batch.Call(map[string]interface{}{"to": opGPOAddr, "input": hexutil.Encode(data)}, "pending", bigIntDecoder, callback)
}

func (batch Batch) EthBalance(account *common.Address, block string, callback func(interface{})) Batch {
	return append(batch, S(bigIntDecoder, callback, "eth_getBalance", account, block))
}

func (batch Batch) GasPrice(callback func(interface{})) Batch {
	return append(batch, S(bigIntDecoder, callback, "eth_gasPrice"))
}

func (batch Batch) ChainId(callback func(interface{})) Batch {
	return append(batch, S(uint64Decoder, callback, "eth_chainID"))
}

func (batch Batch) BlockNumber(callback func(interface{})) Batch {
	return append(batch, S(uint64Decoder, callback, "eth_blockNumber"))
}

func (batch Batch) Nonce(account *common.Address, block string, callback func(interface{})) Batch {
	return append(batch, S(uint64Decoder, callback, "eth_getTransactionCount", account, block))
}

func (batch Batch) SendTx(tx *types.DynamicFeeTx, privateKey *common.Hash, callback func(interface{})) Batch {
	return batch.SendRawTx(utils.SignTx(tx, privateKey), callback)
}

func (batch Batch) ExecuteCall(to *common.Address, call []byte, caller *common.Address, minerTip *big.Int, maxFeePerGas *big.Int, nonce uint64, chainId *big.Int, privateKey *common.Hash, callback func(interface{})) Batch {
	data, err := CallerABI.Pack("execute", to, call)
	if err != nil {
		panic(err)
	}
	return batch.SendTx(&types.DynamicFeeTx{ChainID: chainId, Nonce: nonce, GasTipCap: minerTip, GasFeeCap: maxFeePerGas, Gas: 1000000, To: caller, Value: new(big.Int), Data: data}, privateKey, callback)
}

func (batch Batch) ExecuteTransfer(caller *common.Address, token *common.Address, to *common.Address, amount *big.Int, minerTip *big.Int, maxFeePerGas *big.Int, nonce uint64, chainId *big.Int, privateKey *common.Hash, callback func(interface{})) Batch {
	data, err := Erc20ABI.Pack("transfer", to, amount)
	if err != nil {
		panic(err)
	}
	return batch.ExecuteCall(token, data, caller, minerTip, maxFeePerGas, nonce, chainId, privateKey, callback)
}

func (batch Batch) ExecuteApprove(caller *common.Address, token *common.Address, spender *common.Address, amount *big.Int, minerTip *big.Int, maxFeePerGas *big.Int, nonce uint64, chainId *big.Int, privateKey *common.Hash, callback func(interface{})) Batch {
	data, err := Erc20ABI.Pack("approve", spender, amount)
	if err != nil {
		panic(err)
	}
	return batch.ExecuteCall(token, data, caller, minerTip, maxFeePerGas, nonce, chainId, privateKey, callback)
}

func (batch Batch) SendRawTx(rawTx string, callback func(interface{})) Batch {
	return append(batch, S(nil, callback, "eth_sendRawTransaction", rawTx))
}

func (batch Batch) LogsByTopic(topics [][]string, fromBlock string, toBlock string, callback func(interface{})) Batch {
	return append(batch, S(nil, callback, "eth_getLogs", map[string]interface{}{"fromBlock": fromBlock, "toBlock": toBlock, "topics": topics}))
}

func (batch Batch) BlockByNumber(block string, callback func(interface{})) Batch {
	return append(batch, S(nil, callback, "eth_getBlockByNumber", block, false))
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
