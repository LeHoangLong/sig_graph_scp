package controller_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
	repository_server "sig_graph_scp/pkg/server/repository"

	"github.com/shopspring/decimal"
)

type AssetControllerI interface {
	CreateAsset(
		ctx context.Context,
		user *model_server.User,
		materialName string,
		unit string,
		quantity decimal.Decimal,
		ownerKeyId model_server.UserKeyPairId,
		ingredients []model_server.Asset,
		ingredientSecretIds []string,
		secretIds []string,
		ingredientSignatures []string,
	) (*model_server.Asset, error)
	GetAssetById(ctx context.Context, user *model_server.User, id model_server.NodeId, useCache bool) (*model_server.Asset, error)
	GetOwnedAssetsFromCache(
		ctx context.Context,
		user *model_server.User,
		pagination repository_server.PaginationOption[model_server.NodeDbId],
	) ([]model_server.Asset, error)
}
