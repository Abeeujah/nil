package rawapi

import (
	"context"

	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/internal/execution"
	"github.com/NilFoundation/nil/nil/internal/network"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/txnpool"
)

type LocalShardApi struct {
	db           db.ReadOnlyDB
	accessor     *execution.StateAccessor
	ShardId      types.ShardId
	txnpool      txnpool.Pool
	enableDevApi bool

	nodeApi NodeApi
	logger  logging.Logger
}

var (
	_ ShardApiRo = (*LocalShardApi)(nil)
	_ ShardApi   = (*LocalShardApi)(nil)
)

func NewLocalShardApi(shardId types.ShardId, db db.ReadOnlyDB, txnpool txnpool.Pool, enableDevApi bool) *LocalShardApi {
	stateAccessor := execution.NewStateAccessor()
	return &LocalShardApi{
		db:           db,
		accessor:     stateAccessor,
		ShardId:      shardId,
		txnpool:      txnpool,
		enableDevApi: enableDevApi,
		logger:       logging.NewLogger("local_api"),
	}
}

func (api *LocalShardApi) setAsP2pRequestHandlersIfAllowed(
	ctx context.Context,
	networkManager *network.BasicManager,
	readonly bool,
	logger logging.Logger,
) error {
	return SetRawApiRequestHandlers(ctx, api.ShardId, api, networkManager, readonly, logger)
}

func (api *LocalShardApi) setNodeApi(nodeApi NodeApi) {
	api.nodeApi = nodeApi
}
