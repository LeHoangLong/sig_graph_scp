package service_sig_graph

import (
	"context"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"

	"github.com/shopspring/decimal"
)

type AssetServiceI interface {
	CreateAsset(
		ctx context.Context,
		materialName string,
		unit string,
		quantity decimal.Decimal,
		ownerKey *model_sig_graph.UserKeyPair,
		ingredients []model_sig_graph.Asset,
		ingredientSecretIds []string,
		secretIds []string,
		ingredientSignatures []string,
	) (*model_sig_graph.Asset, error)
	GetAssetById(ctx context.Context, Id string) (*model_sig_graph.Asset, error)

	TransferAsset(
		ctx context.Context,
		time_ms uint64,
		asset *model_sig_graph.Asset,
		newOwnerKey *model_sig_graph.UserKeyPair,
		newId string,
		newSecret string,
		currentSecret string,
		currentSignature string,
	) (*model_sig_graph.Asset, error)
}
