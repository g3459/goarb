package main

import (
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/ethereum/go-ethereum/node"
	"github.com/g3459/goarb/contracts/bytecodes"
	"github.com/g3459/goarb/contracts/interfaces"
)

func main() {
	bh := crypto.Keccak256Hash([]byte{})
	pk := crypto.ToECDSAUnsafe(bh[:])
	sender := crypto.PubkeyToAddress(pk.PublicKey)
	simconf := func(nodeConf *node.Config, ethConf *ethconfig.Config) {
		ethConf.Genesis.Alloc = map[common.Address]types.Account{
			{}:     {Balance: common.MaxHash.Big()},
			sender: {Balance: common.MaxHash.Big()},
		}
		nodeConf.UseLightweightKDF = true
		nodeConf.NoUSB = true
		nodeConf.AllowUnprotectedTxs = true
		nodeConf.JWTSecret = ""
		nodeConf.EnablePersonal = false
		nodeConf.DataDir = ""
		nodeConf.HTTPTimeouts.ReadTimeout = 0
		nodeConf.HTTPTimeouts.WriteTimeout = 0
		nodeConf.HTTPTimeouts.IdleTimeout = 0
		nodeConf.HTTPHost = ""
		nodeConf.HTTPModules = nil
		nodeConf.HTTPCors = nil
		nodeConf.HTTPVirtualHosts = nil
		nodeConf.WSHost = "127.0.0.1"
		nodeConf.WSPort = 8546
		nodeConf.WSModules = []string{"eth", "net", "web3"}
		nodeConf.P2P.NoDiscovery = true
		nodeConf.P2P.ListenAddr = ""
		nodeConf.P2P.MaxPeers = 0
		ethConf.Genesis.GasLimit = 0x7fffffffffffffff
		ethConf.Genesis.BaseFee = new(big.Int)
		ethConf.Genesis.Coinbase = common.Address{}
		ethConf.Genesis.Difficulty = new(big.Int)
		ethConf.Miner.GasCeil = 0x7fffffffffffffff
		ethConf.Miner.GasPrice = new(big.Int)
		ethConf.Miner.Etherbase = common.Address{}
		ethConf.BlobPool.Datadir = ""
		ethConf.BlobPool.Datacap = 0
		ethConf.BlobPool.PriceBump = 0
		ethConf.TxPool.PriceLimit = 0
		ethConf.TxPool.Locals = []common.Address{
			{}, sender,
		}
		ethConf.TxPool.NoLocals = true
		ethConf.GPO.Blocks = 0
		ethConf.GPO.Percentile = 0
		ethConf.GPO.MaxHeaderHistory = 0
		ethConf.GPO.MaxBlockHistory = 0
		ethConf.GPO.MaxPrice = big.NewInt(1)
		ethConf.GPO.IgnorePrice = big.NewInt(1)
		ethConf.GPO.Default = big.NewInt(0)
		ethConf.RPCEVMTimeout = 0
		ethConf.RPCGasCap = 0
		ethConf.RPCTxFeeCap = 0
		ethConf.NoPruning = true
		ethConf.NoPrefetch = false
		ethConf.DatabaseCache = 8192
		ethConf.TrieCleanCache = 4096
		ethConf.TrieDirtyCache = 4096
		ethConf.SnapshotCache = 4096
		ethConf.EnablePreimageRecording = false
		ethConf.VMTrace = ""
		ethConf.VMTraceJsonConfig = ""
		ethConf.TransactionHistory = 0
		ethConf.StateHistory = 0
		ethConf.SkipBcVersionCheck = true
	}
	sim := simulated.NewBackend(nil, simconf)
	auth, err := bind.NewKeyedTransactorWithChainID(pk, big.NewInt(1337))
	if err != nil {
		panic("NewKeyedTransactorWithChainID Err:" + err.Error())
	}
	router, _, _, err := bind.DeployContract(auth, interfaces.RouterABI, bytecodes.RouterBytecode, sim.Client())
	if err != nil {
		panic("DeployContract Err:" + err.Error())
	}
	sim.Commit()
	log.Println("Router:", router)
	select {}
}
