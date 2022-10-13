package api_sig_graph

import (
	"context"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"

	"github.com/shopspring/decimal"
)

type AssetClientApi interface {
	CreateAsset(ctx context.Context, MaterialName string, Unit string, Quantity decimal.Decimal) (*model_sig_graph.Asset, error)
	GetAssetById(ctx context.Context, Id string) (*model_sig_graph.Asset, error)
}

type assetClientApi struct {
}

func (a *assetClientApi) CreateAsset(ctx context.Context, materialName string, unit string, quantity decimal.Decimal, ingredients []model_sig_graph.Asset) (*model_sig_graph.Asset, error) {
	return nil, nil
}

func (a *assetClientApi) GetAssetById(ctx context.Context, Id string) (*model_sig_graph.Asset, error) {
	return nil, nil
}
