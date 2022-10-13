package repository_server

import (
	"context"
	"sig_graph_scp/pkg/model"
)

type AssetRepositoryI interface {
	SaveAsset(ctx context.Context, transactionId TransactionId, asset *model.Asset) error
	FetchAssetsByOwner(ctx context.Context, transactionId TransactionId, namespace string, ownerPublicKey string) ([]model.Asset, error)
	FetchAssetsByIds(ctx context.Context, transactionId TransactionId, namespace string, id map[model.NodeId]bool) ([]model.Asset, error)
}
