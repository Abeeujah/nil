//go:build test

package tests

import (
	"context"
	"fmt"
	"os"
	"slices"
	"sync"

	"github.com/NilFoundation/nil/nil/client"
	rpc_client "github.com/NilFoundation/nil/nil/client/rpc"
	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/internal/abi"
	"github.com/NilFoundation/nil/nil/internal/config"
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/internal/execution"
	"github.com/NilFoundation/nil/nil/internal/keys"
	"github.com/NilFoundation/nil/nil/internal/network"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/nilservice"
	"github.com/NilFoundation/nil/nil/services/rpc"
	"github.com/NilFoundation/nil/nil/services/rpc/jsonrpc"
	"github.com/rs/zerolog"
)

type Shard struct {
	Id         types.ShardId
	Db         db.DB
	RpcUrl     string
	P2pAddress network.AddrInfo
	Client     client.Client
	nm         *network.Manager
	Config     *nilservice.Config
}

func getShardAddress(s Shard) network.AddrInfo {
	return s.P2pAddress
}

type ShardedSuite struct {
	CliRunner

	DefaultClient client.Client
	Context       context.Context
	ctxCancel     context.CancelFunc
	Wg            sync.WaitGroup

	dbInit func() db.DB

	Shards []Shard
}

type DhtBootstrapByValidators int

const (
	WithoutDhtBootstrapByValidators DhtBootstrapByValidators = iota
	WithDhtBootstrapByValidators
)

func (s *ShardedSuite) Cancel() {
	s.T().Helper()

	s.ctxCancel()
	s.Wg.Wait()
	for _, shard := range s.Shards {
		shard.Db.Close()
	}
}

func (s *ShardedSuite) createOneShardOneValidatorCfg(
	shardId types.ShardId, cfg *nilservice.Config, netCfg *network.Config, keyManagers map[types.ShardId]*keys.ValidatorKeysManager,
) *nilservice.Config {
	validatorKeysPath := keyManagers[shardId].GetKeysPath()

	validators := make(map[types.ShardId][]config.ValidatorInfo)
	for kmShardId, km := range keyManagers {
		pkey, err := km.GetPublicKey(kmShardId)
		s.Require().NoError(err)
		validators[kmShardId] = []config.ValidatorInfo{
			{PublicKey: [33]byte(pkey)},
		}
	}

	s.Require().NotEmpty(validatorKeysPath)
	return &nilservice.Config{
		NShards:              cfg.NShards,
		MyShards:             []uint{uint(shardId)},
		SplitShards:          true,
		HttpUrl:              s.Shards[shardId].RpcUrl,
		Topology:             cfg.Topology,
		CollatorTickPeriodMs: cfg.CollatorTickPeriodMs,
		GasBasePrice:         cfg.GasBasePrice,
		Network:              netCfg,
		ZeroStateYaml:        cfg.ZeroStateYaml,
		ValidatorKeysPath:    validatorKeysPath,
		Validators:           validators,
	}
}

func (s *ShardedSuite) start(cfg *nilservice.Config, port int) {
	s.T().Helper()
	s.Context, s.ctxCancel = context.WithCancel(context.Background())

	if s.dbInit == nil {
		s.dbInit = func() db.DB {
			db, err := db.NewBadgerDbInMemory()
			s.Require().NoError(err)
			return db
		}
	}

	networkConfigs, p2pAddresses := network.GenerateConfigs(s.T(), cfg.NShards, port)

	keysManagers := make(map[types.ShardId]*keys.ValidatorKeysManager)
	s.Shards = make([]Shard, 0, cfg.NShards)
	for i := range cfg.NShards {
		shardId := types.ShardId(i)

		keysPath := s.T().TempDir() + fmt.Sprintf("/validator-keys-%d.yaml", i)
		km := keys.NewValidatorKeyManager(keysPath, cfg.NShards)
		s.Require().NotNil(km)
		s.Require().NoError(km.InitKeys())
		keysManagers[shardId] = km

		url := rpc.GetSockPathIdx(s.T(), int(i))
		shard := Shard{
			Id:         shardId,
			Db:         s.dbInit(),
			RpcUrl:     url,
			P2pAddress: p2pAddresses[i],
		}
		shard.Client = rpc_client.NewClient(shard.RpcUrl, zerolog.New(os.Stderr))
		s.Shards = append(s.Shards, shard)
	}

	PatchConfigWithTestDefaults(cfg)
	for i := range types.ShardId(cfg.NShards) {
		shardConfig := s.createOneShardOneValidatorCfg(i, cfg, networkConfigs[i], keysManagers)

		node, err := nilservice.CreateNode(s.Context, fmt.Sprintf("shard-%d", i), shardConfig, s.Shards[i].Db, nil)
		s.Require().NoError(err)
		s.Shards[i].nm = node.NetworkManager
		s.Shards[i].Config = shardConfig

		s.Wg.Add(1)
		go func() {
			defer s.Wg.Done()
			defer node.Close(s.Context)
			s.NoError(node.Run())
		}()
	}

	for _, shard := range s.Shards {
		s.connectToShards(shard.nm)
	}

	s.waitZerostate()
}

func (s *ShardedSuite) Start(cfg *nilservice.Config, port int) {
	s.T().Helper()

	s.start(cfg, port)
}

func (s *ShardedSuite) connectToShards(nm *network.Manager) {
	s.T().Helper()

	var wg sync.WaitGroup
	for _, shard := range s.Shards {
		if shard.nm != nm {
			wg.Add(1)
			go func() {
				defer wg.Done()
				network.ConnectManagers(s.T(), nm, shard.nm)
			}()
		}
	}
	wg.Wait()
}

func (s *ShardedSuite) StartArchiveNode(port int, withBootstrapPeers bool) (client.Client, network.AddrInfo) {
	s.T().Helper()

	s.Require().NotEmpty(s.Shards)
	netCfg, addr := network.GenerateConfig(s.T(), port)
	serviceName := fmt.Sprintf("archive-%d", port)

	cfg := &nilservice.Config{
		NShards:    uint32(len(s.Shards)),
		Network:    netCfg,
		HttpUrl:    rpc.GetSockPathService(s.T(), serviceName),
		RunMode:    nilservice.ArchiveRunMode,
		Validators: s.Shards[0].Config.Validators,
	}

	cfg.MyShards = slices.Collect(common.Range(0, uint(cfg.NShards)))
	netCfg.DHTBootstrapPeers = slices.Collect(common.Transform(slices.Values(s.Shards), getShardAddress))
	if withBootstrapPeers {
		cfg.BootstrapPeers = netCfg.DHTBootstrapPeers
	}

	node, err := nilservice.CreateNode(s.Context, serviceName, cfg, s.dbInit(), nil)
	s.Require().NoError(err)
	s.connectToShards(node.NetworkManager)

	s.Wg.Add(1)
	go func() {
		defer s.Wg.Done()
		defer node.Close(s.Context)
		s.NoError(node.Run())
	}()

	c := rpc_client.NewClient(cfg.HttpUrl, zerolog.New(os.Stderr))
	s.checkNodeStart(cfg.NShards, c)
	return c, addr
}

func (s *ShardedSuite) StartRPCNode(dhtBootstrapByValidators DhtBootstrapByValidators, archiveNodes network.AddrInfoSlice) (client.Client, string) {
	s.T().Helper()

	netCfg, _ := network.GenerateConfig(s.T(), 0)
	const serviceName = "rpc"

	cfg := &nilservice.Config{
		NShards: uint32(len(s.Shards)),
		Network: netCfg,
		HttpUrl: rpc.GetSockPathService(s.T(), serviceName),
		RunMode: nilservice.RpcRunMode,
		RpcNode: nilservice.NewDefaultRpcNodeConfig(),
	}

	if dhtBootstrapByValidators == WithDhtBootstrapByValidators {
		netCfg.DHTBootstrapPeers = slices.Collect(common.Transform(slices.Values(s.Shards), getShardAddress))
	}
	cfg.RpcNode.ArchiveNodeList = archiveNodes

	node, err := nilservice.CreateNode(s.Context, serviceName, cfg, s.dbInit(), nil)
	s.Require().NoError(err)
	if dhtBootstrapByValidators == WithDhtBootstrapByValidators {
		s.connectToShards(node.NetworkManager)
	}

	s.Wg.Add(1)
	go func() {
		defer s.Wg.Done()
		defer node.Close(s.Context)
		s.NoError(node.Run())
	}()

	c := rpc_client.NewClient(cfg.HttpUrl, zerolog.New(os.Stderr))
	s.checkNodeStart(cfg.NShards, c)
	return c, cfg.HttpUrl
}

func (s *ShardedSuite) WaitForReceipt(hash common.Hash) *jsonrpc.RPCReceipt {
	s.T().Helper()

	return WaitForReceipt(s.T(), s.Context, s.DefaultClient, hash)
}

func (s *ShardedSuite) WaitIncludedInMain(hash common.Hash) *jsonrpc.RPCReceipt {
	s.T().Helper()

	return WaitIncludedInMain(s.T(), s.Context, s.DefaultClient, hash)
}

func (s *ShardedSuite) GasToValue(gas uint64) types.Value {
	return GasToValue(gas)
}

func (s *ShardedSuite) DeployContractViaMainSmartAccount(shardId types.ShardId, payload types.DeployPayload, initialAmount types.Value) (types.Address, *jsonrpc.RPCReceipt) {
	s.T().Helper()

	return DeployContractViaSmartAccount(s.T(), s.Context, s.DefaultClient, types.MainSmartAccountAddress, execution.MainPrivateKey, shardId, payload, initialAmount)
}

func (s *ShardedSuite) checkNodeStart(nShards uint32, client client.Client) {
	s.T().Helper()

	var wg sync.WaitGroup
	wg.Add(int(nShards))
	for shardId := range types.ShardId(nShards) {
		go func() {
			defer wg.Done()
			WaitZerostate(s.T(), s.Context, client, shardId)
		}()
	}
	wg.Wait()
}

func (s *ShardedSuite) waitZerostate() {
	s.T().Helper()

	var wg sync.WaitGroup
	wg.Add(len(s.Shards))
	for _, shard := range s.Shards {
		go func() {
			defer wg.Done()
			WaitZerostate(s.T(), s.Context, shard.Client, shard.Id)
		}()
	}
	wg.Wait()
}

func (s *ShardedSuite) LoadContract(path string, name string) (types.Code, abi.ABI) {
	s.T().Helper()
	return LoadContract(s.T(), path, name)
}

func (s *ShardedSuite) PrepareDefaultDeployPayload(abi abi.ABI, code []byte, args ...any) types.DeployPayload {
	s.T().Helper()
	return PrepareDefaultDeployPayload(s.T(), abi, code, args...)
}

func (s *ShardedSuite) GetBalance(address types.Address) types.Value {
	s.T().Helper()
	return GetBalance(s.T(), s.Context, s.DefaultClient, address)
}

func (s *ShardedSuite) AbiPack(abi *abi.ABI, name string, args ...any) []byte {
	s.T().Helper()
	return AbiPack(s.T(), abi, name, args...)
}

func (s *ShardedSuite) SendExternalTransactionNoCheck(bytecode types.Code, contractAddress types.Address) *jsonrpc.RPCReceipt {
	s.T().Helper()
	return SendExternalTransactionNoCheck(s.T(), s.Context, s.DefaultClient, bytecode, contractAddress)
}

func (s *ShardedSuite) AnalyzeReceipt(receipt *jsonrpc.RPCReceipt, namesMap map[types.Address]string) ReceiptInfo {
	s.T().Helper()
	return AnalyzeReceipt(s.T(), s.Context, s.DefaultClient, receipt, namesMap)
}

func (s *ShardedSuite) CheckBalance(infoMap ReceiptInfo, balance types.Value, accounts []types.Address) types.Value {
	s.T().Helper()
	return CheckBalance(s.T(), s.Context, s.DefaultClient, infoMap, balance, accounts)
}

func (s *ShardedSuite) CallGetter(addr types.Address, calldata []byte, blockId any, overrides *jsonrpc.StateOverrides) []byte {
	s.T().Helper()
	return CallGetter(s.T(), s.Context, s.DefaultClient, addr, calldata, blockId, overrides)
}
