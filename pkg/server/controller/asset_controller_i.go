package controller_server

import (
	"context"
	"sig_graph_scp/pkg/model"

	"github.com/shopspring/decimal"
)

type AssetControllerI interface {
	CreateAsset(
		ctx context.Context,
		user *model.User,
		materialName string,
		unit string,
		quantity decimal.Decimal,
		ownerKeyId model.UserKeyPairId,
		ingredients []model.Asset,
		ingredientSecretIds []string,
		secretIds []string,
		ingredientSignatures []string,
	) (*model.Asset, error)
	GetAssetById(ctx context.Context, user *model.User, id model.NodeId, useCache bool) (*model.Asset, error)
	GetOwnedAssetsFromCache(ctx context.Context, user *model.User) ([]model.Asset, error)
}
