package simulated

import (
	"errors"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/eth/catalyst"
	"github.com/ethereum/go-ethereum/eth/downloader"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/eth/filters"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
)

type Backend struct {
	node   *node.Node
	beacon *catalyst.SimulatedBeacon
	client *ethclient.Client
}

func NewSimulated() *Backend {
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
	sim := newBackend(nil, simconf)
	return sim
}

func newBackend(alloc types.GenesisAlloc, options ...func(nodeConf *node.Config, ethConf *ethconfig.Config)) *Backend {
	// Create the default configurations for the outer node shell and the Ethereum
	// service to mutate with the options afterwards
	nodeConf := node.DefaultConfig
	nodeConf.DataDir = ""
	nodeConf.P2P = p2p.Config{NoDiscovery: true}

	ethConf := ethconfig.Defaults
	ethConf.Genesis = &core.Genesis{
		Config:   params.AllDevChainProtocolChanges,
		GasLimit: ethconfig.Defaults.Miner.GasCeil,
		Alloc:    alloc,
	}
	ethConf.SyncMode = downloader.FullSync
	ethConf.TxPool.NoLocals = true

	for _, option := range options {
		option(&nodeConf, &ethConf)
	}
	// Assemble the Ethereum stack to run the chain with
	stack, err := node.New(&nodeConf)
	if err != nil {
		panic(err) // this should never happen
	}
	sim, err := newWithNode(stack, &ethConf, 0)
	if err != nil {
		panic(err) // this should never happen
	}
	return sim
}

func newWithNode(stack *node.Node, conf *eth.Config, blockPeriod uint64) (*Backend, error) {
	backend, err := eth.New(stack, conf)
	if err != nil {
		return nil, err
	}
	// Register the filter system
	filterSystem := filters.NewFilterSystem(backend.APIBackend, filters.Config{})
	stack.RegisterAPIs([]rpc.API{{
		Namespace: "eth",
		Service:   filters.NewFilterAPI(filterSystem),
	}})
	// Start the node
	if err := stack.Start(); err != nil {
		return nil, err
	}
	// Set up the simulated beacon
	beacon, err := catalyst.NewSimulatedBeacon(blockPeriod, backend)
	if err != nil {
		return nil, err
	}
	// Reorg our chain back to genesis
	if err := beacon.Fork(backend.BlockChain().GetCanonicalHash(0)); err != nil {
		return nil, err
	}
	return &Backend{
		node:   stack,
		beacon: beacon,
		client: ethclient.NewClient(stack.Attach()),
	}, nil
}

func DeployContract(sim *Backend, bytecode []byte) (*common.Address, error) {
	bh := crypto.Keccak256Hash([]byte{})
	pk := crypto.ToECDSAUnsafe(bh[:])
	auth, err := bind.NewKeyedTransactorWithChainID(pk, big.NewInt(1337))
	if err != nil {
		return nil, errors.New("NewKeyedTransactorWithChainID Err:" + err.Error())
	}
	router, _, _, err := bind.DeployContract(auth, abi.ABI{}, bytecode, sim.Client())
	if err != nil {
		return nil, errors.New("DeployContract Err:" + err.Error())
	}
	sim.Commit()
	return &router, nil
}

// Close shuts down the simBackend.
// The simulated backend can't be used afterwards.
func (n *Backend) Close() error {
	if n.client != nil {
		n.client.Close()
		n.client = &ethclient.Client{}
	}
	var err error
	if n.beacon != nil {
		err = n.beacon.Stop()
		n.beacon = nil
	}
	if n.node != nil {
		err = errors.Join(err, n.node.Close())
		n.node = nil
	}
	return err
}

// Commit seals a block and moves the chain forward to a new empty block.
func (n *Backend) Commit() common.Hash {
	return n.beacon.Commit()
}

// Rollback removes all pending transactions, reverting to the last committed state.
func (n *Backend) Rollback() {
	n.beacon.Rollback()
}

// Fork creates a side-chain that can be used to simulate reorgs.
//
// This function should be called with the ancestor block where the new side
// chain should be started. Transactions (old and new) can then be applied on
// top and Commit-ed.
//
// Note, the side-chain will only become canonical (and trigger the events) when
// it becomes longer. Until then CallContract will still operate on the current
// canonical chain.
//
// There is a % chance that the side chain becomes canonical at the same length
// to simulate live network behavior.
func (n *Backend) Fork(parentHash common.Hash) error {
	return n.beacon.Fork(parentHash)
}

// AdjustTime changes the block timestamp and creates a new block.
// It can only be called on empty blocks.
func (n *Backend) AdjustTime(adjustment time.Duration) error {
	return n.beacon.AdjustTime(adjustment)
}

// Client returns a client that accesses the simulated chain.
func (n *Backend) Client() *ethclient.Client {
	return n.client
}
