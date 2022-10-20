package service_sig_graph

import (
	"context"
	"encoding/json"
	"sig_graph_scp/pkg/model"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"
	"sig_graph_scp/pkg/utility"

	"github.com/shopspring/decimal"
)

type assetService struct {
	smartContractService SmartContractServiceI
	clock                utility.ClockI
	idGenerateService    IdGenerateServiceI
	signingService       NodeSigningServiceI
}

func NewAssetService(
	smartContractService SmartContractServiceI,
	clock utility.ClockI,
	idGenerateService IdGenerateServiceI,
	signingService NodeSigningServiceI,
) AssetServiceI {
	return &assetService{
		smartContractService: smartContractService,
		clock:                clock,
		idGenerateService:    idGenerateService,
		signingService:       signingService,
	}
}

type createAssetRequest struct {
	Time                 uint64   `json:"time"`
	Id                   string   `json:"id"`
	MaterialName         string   `json:"material_name"`
	Quantity             string   `json:"quantity"`
	Unit                 string   `json:"unit"`
	Signature            string   `json:"signature"`
	OwnerPublicKey       string   `json:"owner_public_key"`
	IngredientIds        []string `json:"ingredient_ids"`
	IngredientSecretIds  []string `json:"ingredient_secret_ids"`
	SecretIds            []string `json:"secret_ids"`
	IngredientSignatures []string `json:"ingredient_signatures"`
}

func (s *assetService) CreateAsset(
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
	ingredientIds := []string{}
	for i := range ingredients {
		ingredientIds = append(ingredientIds, string(ingredients[i].Id))
	}

	time := s.clock.Now()
	time_ms := time.UnixMilli()
	id := s.idGenerateService.NewFullId()

	// generate signature
	node := model_sig_graph.NewDefaultNode(
		id,
		model.ENodeTypeAsset,
		uint64(time_ms),
		uint64(time_ms),
		"",
		ownerKey.Public,
	)
	asset := model_sig_graph.NewAsset(
		node,
		model.ECreationProcessCreate,
		unit,
		quantity,
		materialName,
	)

	signature, err := s.signingService.Sign(ctx, ownerKey, asset)
	if err != nil {
		return nil, err
	}

	request := createAssetRequest{
		Time:                 uint64(time_ms),
		Id:                   string(id),
		MaterialName:         materialName,
		Quantity:             quantity.String(),
		Unit:                 unit,
		Signature:            signature,
		OwnerPublicKey:       ownerKey.Public,
		IngredientIds:        ingredientIds,
		IngredientSecretIds:  ingredientSecretIds,
		SecretIds:            secretIds,
		IngredientSignatures: ingredientSignatures,
	}

	requestJson, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	assetStr, err := s.smartContractService.CreateTransaction("CreateAsset", string(requestJson))
	if err != nil {
		return nil, err
	}

	assetSigGraph := model_sig_graph.Asset{}
	err = json.Unmarshal([]byte(assetStr), &assetSigGraph)
	if err != nil {
		return nil, err
	}
	return &assetSigGraph, nil
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
