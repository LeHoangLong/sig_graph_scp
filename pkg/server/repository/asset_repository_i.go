package repository_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
)

type AssetRepositoryI interface {
	SaveAsset(ctx context.Context, transactionId TransactionId, asset *model_server.Asset) error
	FetchAssetsByOwner(ctx context.Context, transactionId TransactionId, namespace string, ownerPublicKey string, finalizeStatus []bool, pagination PaginationOption[model_server.NodeDbId]) ([]model_server.Asset, error)
	FetchAssetsByIds(ctx context.Context, transactionId TransactionId, namespace string, id map[model_server.NodeId]bool) ([]model_server.Asset, error)
	FetchAssetsByDbIds(ctx context.Context, transactionId TransactionId, namespace string, id map[model_server.NodeDbId]bool) ([]model_server.Asset, error)
}
