package service_sig_graph

import (
	"context"
	"sig_graph_scp/pkg/model"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"

	"github.com/shopspring/decimal"
)

type AssetServiceI interface {
	CreateAsset(
		ctx context.Context,
		materialName string,
		unit string,
		quantity decimal.Decimal,
		ownerKey *model.UserKeyPair,
		ingredients []model_sig_graph.Asset,
		ingredientSecretIds []string,
		secretIds []string,
		ingredientSignatures []string,
	) (*model_sig_graph.Asset, error)
	GetAssetById(ctx context.Context, Id model.NodeId) (*model_sig_graph.Asset, error)
}
