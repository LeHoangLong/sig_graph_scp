package controller_server

import (
	"context"
	"sig_graph_scp/pkg/model"
)

type AssetControllerI interface {
	GetAssetById(ctx context.Context, user *model.User, id model.NodeId, useCache bool) (*model.Asset, error)
	GetOwnedAssetsFromCache(ctx context.Context, user *model.User) ([]model.Asset, error)
}
