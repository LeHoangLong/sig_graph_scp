package api_sig_graph

import (
	"context"
	"fmt"
	service_sig_graph "sig_graph_scp/internal/sig_graph/service"
	"sig_graph_scp/pkg/model"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"
	"sig_graph_scp/pkg/utility"

	"github.com/shopspring/decimal"
)

type AssetClientApi interface {
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

type assetClientApi struct {
	assetService service_sig_graph.AssetServiceI
}

func NewAssetClientApi(graphName string) (AssetClientApi, error) {
	smartContractService, err := service_sig_graph.NewSmartContractServiceHyperledger()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize smart contract service: %w", err)
	}

	nodeSigningService := service_sig_graph.NewNodeSigningService()
	idGeneratorService := service_sig_graph.NewIdGenerateServiceUuid(graphName)
	clockWall := utility.NewClockWall()

	assetSigGraphService := service_sig_graph.NewAssetService(smartContractService, clockWall, idGeneratorService, nodeSigningService)
	return &assetClientApi{
		assetService: assetSigGraphService,
	}, nil
}

func (a *assetClientApi) CreateAsset(
	ctx context.Context,
	materialName string,
	unit string,
	quantity decimal.Decimal,
	ownerKey *model.UserKeyPair,
	ingredients []model_sig_graph.Asset,
	ingredientSecretIds []string,
	secretIds []string,
	ingredientSignatures []string,
) (*model_sig_graph.Asset, error) {
	return a.assetService.CreateAsset(
		ctx,
		materialName,
		unit,
		quantity,
		ownerKey,
		ingredients,
		ingredientSecretIds,
		secretIds,
		ingredientSignatures,
	)
}

func (a *assetClientApi) GetAssetById(ctx context.Context, Id model.NodeId) (*model_sig_graph.Asset, error) {
	return a.assetService.GetAssetById(ctx, Id)
}
