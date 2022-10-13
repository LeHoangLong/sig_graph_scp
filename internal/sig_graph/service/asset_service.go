package service_sig_graph

import (
	"context"
	"encoding/json"
	"sig_graph_scp/pkg/model"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"

	"github.com/shopspring/decimal"
)

type assetService struct {
	smartContractService SmartContractServiceI
}

func NewAssetService(
	smartContractService SmartContractServiceI,
) AssetServiceI {
	return &assetService{
		smartContractService: smartContractService,
	}
}

func (s *assetService) CreateAsset(
	ctx context.Context,
	MaterialName string,
	Unit string,
	Quantity decimal.Decimal,
	ingredients []model_sig_graph.Asset,
) (*model_sig_graph.Asset, error) {
	return nil, nil
}

func (s *assetService) GetAssetById(ctx context.Context, id model.NodeId) (*model_sig_graph.Asset, error) {
	data, err := s.smartContractService.Query("GetAsset", string(id))
	if err != nil {
		return nil, err
	}

	asset := model_sig_graph.Asset{}
	err = json.Unmarshal([]byte(data), &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}
