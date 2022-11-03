package api_sig_graph

import (
	"context"
	"fmt"
	service_sig_graph "sig_graph_scp/internal/sig_graph/service"
	model_server "sig_graph_scp/pkg/server/model"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"
	"sig_graph_scp/pkg/utility"

	"github.com/shopspring/decimal"
)

type SigGraphClientApi interface {
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
	GetAssetById(ctx context.Context, Id model_server.NodeId) (*model_sig_graph.Asset, error)
	DoNodeIdsExists(ctx context.Context, ids map[string]bool) (map[string]bool, error)
	TransferAsset(
		ctx context.Context,
		time_ms uint64,
		asset *model_sig_graph.Asset,
		newOwnerKey *model_sig_graph.UserKeyPair,
		newId string,
		newSecret string,
		currentSecret string,
		currentSignature string,
	) (updatedCurrentAsset *model_sig_graph.Asset, newAsset *model_sig_graph.Asset, err error)
	GetGraphName() string
}

type sigGraphClientApi struct {
	assetService service_sig_graph.AssetServiceI
	nodeService  service_sig_graph.NodeServiceI
	graphName    string
}

func NewAssetClientApi(graphName string) (SigGraphClientApi, error) {
	smartContractService, err := service_sig_graph.NewSmartContractServiceHyperledger()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize smart contract service: %w", err)
	}

	nodeSigningService := service_sig_graph.NewNodeSigningService()
	idGeneratorService := service_sig_graph.NewIdGenerateServiceUuid(graphName)
	clockWall := utility.NewClockWall()
	hashGenerator := utility.NewHashedIdGeneratorService()
	cloner := utility.NewCloner()

	nodeSigGraphService := service_sig_graph.NewNodeService(smartContractService)
	assetSigGraphService := service_sig_graph.NewAssetService(
		smartContractService,
		clockWall,
		idGeneratorService,
		nodeSigningService,
		hashGenerator,
		cloner,
	)
	return &sigGraphClientApi{
		assetService: assetSigGraphService,
		nodeService:  nodeSigGraphService,
		graphName:    graphName,
	}, nil
}

func (a *sigGraphClientApi) GetGraphName() string {
	return a.graphName
}

func (a *sigGraphClientApi) DoNodeIdsExists(ctx context.Context, ids map[string]bool) (map[string]bool, error) {
	return a.nodeService.DoNodeIdsExists(ctx, ids)
}

func (a *sigGraphClientApi) CreateAsset(
	ctx context.Context,
	materialName string,
	unit string,
	quantity decimal.Decimal,
	ownerKey *model_sig_graph.UserKeyPair,
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

func (a *sigGraphClientApi) GetAssetById(ctx context.Context, Id model_server.NodeId) (*model_sig_graph.Asset, error) {
	return a.assetService.GetAssetById(ctx, string(Id))
}

func (a *sigGraphClientApi) TransferAsset(
	ctx context.Context,
	time_ms uint64,
	asset *model_sig_graph.Asset,
	newOwnerKey *model_sig_graph.UserKeyPair,
	newId string,
	newSecret string,
	currentSecret string,
	currentSignature string,
) (updatedCurrentAsset *model_sig_graph.Asset, newAsset *model_sig_graph.Asset, err error) {
	return a.assetService.TransferAsset(
		ctx,
		time_ms,
		asset,
		newOwnerKey,
		newId,
		newSecret,
		currentSecret,
		currentSignature,
	)
}
