package service_sig_graph

import (
	"context"
	"crypto/sha512"
	"encoding/json"
	"fmt"
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
	ownerKey *model_sig_graph.UserKeyPair,
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
	id, err := s.idGenerateService.NewFullId(ctx)
	if err != nil {
		return nil, err
	}

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

func (s *assetService) GetAssetById(ctx context.Context, id string) (*model_sig_graph.Asset, error) {
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

type transefrAssetRequest struct {
	TimeMs uint64 `json:"time_ms"`

	CurrentId        string `json:"current_id"`
	CurrentSignature string `json:"current_signature"`
	CurrentSecret    string `json:"current_secret"`

	NewId        string `json:"new_id"`
	NewSignature string `json:"new_signature"`
	NewSecret    string `json:"new_secret"`

	NewOwnerPublicKey string `json:"new_owner_public_key"`
}

func (s *assetService) TransferAsset(
	ctx context.Context,
	time_ms uint64,
	asset *model_sig_graph.Asset,
	newOwnerKey *model_sig_graph.UserKeyPair,
	newId string,
	newSecret string,
	currentSecret string,
	currentSignature string,
) (*model_sig_graph.Asset, error) {
	newAsset := *asset
	newAsset.Id = newId
	newAsset.OwnerPublicKey = newOwnerKey.Public
	if newSecret != "" {
		currentSecretId := fmt.Sprintf("%s%s", asset.Id, currentSecret)
		currentHashBytes := sha512.Sum512([]byte(currentSecretId))
		currentHash := string(currentHashBytes[:])
		newAsset.PrivateParentsHashedIds[currentHash] = true
	} else {
		newAsset.PublicParentsIds[asset.Id] = true
	}

	signature, err := s.signingService.Sign(
		ctx,
		newOwnerKey,
		newAsset,
	)

	if err != nil {
		return nil, err
	}

	request := transefrAssetRequest{
		TimeMs:           time_ms,
		CurrentId:        asset.Id,
		CurrentSignature: currentSignature,
		CurrentSecret:    currentSecret,

		NewId:        newId,
		NewSignature: signature,
		NewSecret:    newSecret,

		NewOwnerPublicKey: newOwnerKey.Public,
	}

	requestJson, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	assetStr, err := s.smartContractService.CreateTransaction("TransferAsset", string(requestJson))
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
