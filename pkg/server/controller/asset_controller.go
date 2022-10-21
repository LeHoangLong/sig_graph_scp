package controller_server

import (
	"context"
	"crypto/sha512"
	"fmt"
	"sig_graph_scp/pkg/model"
	repository_server "sig_graph_scp/pkg/server/repository"
	api_sig_graph "sig_graph_scp/pkg/sig_graph/api"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"
	"sig_graph_scp/pkg/utility"

	"github.com/shopspring/decimal"
)

type assetController struct {
	api                api_sig_graph.AssetClientApi
	repository         repository_server.AssetRepositoryI
	keyRepository      repository_server.UserKeyRepositoryI
	transactionManager repository_server.TransactionManagerI
}

func NewAssetController(
	api api_sig_graph.AssetClientApi,
	repository repository_server.AssetRepositoryI,
	keyRepository repository_server.UserKeyRepositoryI,
	transactionManager repository_server.TransactionManagerI,
) AssetControllerI {
	return &assetController{
		api:                api,
		repository:         repository,
		keyRepository:      keyRepository,
		transactionManager: transactionManager,
	}
}

func (c *assetController) CreateAsset(
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
) (*model.Asset, error) {
	if len(ingredients) != len(ingredientSecretIds) {
		return nil, fmt.Errorf("mismatch length")
	}

	if len(ingredients) != len(secretIds) {
		return nil, fmt.Errorf("mismatch length")
	}

	if len(ingredients) != len(ingredientSignatures) {
		return nil, fmt.Errorf("mismatch length")
	}

	transactionId, err := c.transactionManager.BypassTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer c.transactionManager.StopBypassedTransaction(ctx, transactionId)

	ownerKeys, err := c.keyRepository.FetchKeyPairsOfUser(ctx, transactionId, user)
	if err != nil {
		return nil, err
	}

	var ownerKey *model.UserKeyPair = nil
	for i := range ownerKeys {
		if ownerKeys[i].Id == ownerKeyId {
			ownerKey = &ownerKeys[i]
		}
	}

	if ownerKey == nil {
		return nil, fmt.Errorf("%w: no owner key with id %d", utility.ErrNotFound, ownerKeyId)
	}

	sigGraphIngredients := make([]model_sig_graph.Asset, 0, len(ingredients))
	for i := range ingredients {
		sigGraphIngredients = append(sigGraphIngredients, model_sig_graph.FromModelAsset(&ingredients[i]))
	}

	asset, err := c.api.CreateAsset(
		ctx,
		materialName,
		unit,
		quantity,
		ownerKey,
		sigGraphIngredients,
		ingredientSecretIds,
		secretIds,
		ingredientSignatures,
	)

	if err != nil {
		return nil, err
	}

	secretParentIds := map[string]model.PrivateId{}
	for i := range ingredients {
		thisId := ingredients[i].Id
		thisSecret := ingredientSecretIds[i]

		if thisSecret != "" {
			thisSecretId := fmt.Sprintf("%s%s", thisId, thisSecret)
			thisSum := sha512.Sum512([]byte(thisSecretId))
			thisHash := string(thisSum[:])

			otherId := asset.Id
			otherSecret := secretIds[i]
			otherSecretId := fmt.Sprintf("%s%s", otherId, otherSecret)
			otherSum := sha512.Sum512([]byte(otherSecretId))
			otherHash := string(otherSum[:])

			privateId := model.PrivateId{
				ThisId:     thisId,
				ThisHash:   thisHash,
				ThisSecret: thisSecret,

				OtherId:     otherId,
				OtherSecret: otherSecret,
				OtherHash:   otherHash,
			}

			secretParentIds[thisHash] = privateId
		}
	}

	namespace := fmt.Sprintf("%d", user.ID)
	modelNode := model_sig_graph.ToModelNode(&asset.Node, 0, namespace, secretParentIds, map[string]model.PrivateId{})
	modelAsset := model_sig_graph.ToModelAsset(asset, &modelNode)

	// save asset
	c.repository.SaveAsset(ctx, transactionId, &modelAsset)

	return &modelAsset, nil
}

func (c *assetController) GetAssetById(ctx context.Context, user *model.User, id model.NodeId, useCache bool) (*model.Asset, error) {
	var transactionId repository_server.TransactionId
	var err error
	if useCache {
		transactionId, err = c.transactionManager.BypassTransaction(ctx)
		if err != nil {
			return nil, err
		}
		defer c.transactionManager.StopBypassedTransaction(ctx, transactionId)

		assets, err := c.repository.FetchAssetsByIds(ctx, transactionId, fmt.Sprintf("%d", user.ID), map[model.NodeId]bool{id: true})
		if err != nil {
			return nil, err
		}

		if len(assets) > 0 {
			return &assets[0], nil
		}
	}

	asset, err := c.api.GetAssetById(ctx, id)
	if err != nil {
		return nil, err
	}

	modelNode := model_sig_graph.ToModelNode(&asset.Node, 0, fmt.Sprintf("%d", user.ID), map[string]model.PrivateId{}, map[string]model.PrivateId{})
	modelAsset := model_sig_graph.ToModelAsset(asset, &modelNode)

	c.repository.SaveAsset(ctx, transactionId, &modelAsset)

	return &modelAsset, nil
}

func (c *assetController) GetOwnedAssetsFromCache(ctx context.Context, user *model.User) ([]model.Asset, error) {
	transactionId, err := c.transactionManager.BypassTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer c.transactionManager.StopBypassedTransaction(ctx, transactionId)

	keys, err := c.keyRepository.FetchKeyPairsOfUser(ctx, transactionId, user)
	if err != nil {
		return nil, err
	}

	ret := []model.Asset{}

	for _, key := range keys {
		assets, err := c.repository.FetchAssetsByOwner(ctx, transactionId, fmt.Sprintf("%d", user.ID), key.Public)
		if err != nil {
			return nil, err
		}
		ret = append(ret, assets...)
	}

	return ret, nil
}
