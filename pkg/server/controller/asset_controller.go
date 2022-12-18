package controller_server

import (
	"context"
	"fmt"
	"math"
	model_server "sig_graph_scp/pkg/server/model"
	repository_server "sig_graph_scp/pkg/server/repository"
	api_sig_graph "sig_graph_scp/pkg/sig_graph/api"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"
	"sig_graph_scp/pkg/utility"

	"github.com/shopspring/decimal"
)

type assetController struct {
	api                  api_sig_graph.SigGraphClientApi
	repository           repository_server.AssetRepositoryI
	keyRepository        repository_server.UserKeyRepositoryI
	transactionManager   repository_server.TransactionManagerI
	hashGeneratorService utility.HashedIdGeneratorServiceI
}

func NewAssetController(
	api api_sig_graph.SigGraphClientApi,
	repository repository_server.AssetRepositoryI,
	keyRepository repository_server.UserKeyRepositoryI,
	transactionManager repository_server.TransactionManagerI,
	hashGeneratorService utility.HashedIdGeneratorServiceI,
) AssetControllerI {
	return &assetController{
		api:                  api,
		repository:           repository,
		keyRepository:        keyRepository,
		transactionManager:   transactionManager,
		hashGeneratorService: hashGeneratorService,
	}
}

func (c *assetController) CreateAsset(
	ctx context.Context,
	user *model_server.User,
	materialName string,
	unit string,
	quantity decimal.Decimal,
	ownerKeyId model_server.UserKeyPairId,
	ingredients []model_server.Asset,
	ingredientSecretIds []string,
	secretIds []string,
	ingredientSignatures []string,
) (*model_server.Asset, error) {
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

	keyPagination := repository_server.PaginationOption[model_server.UserKeyPairId]{
		MinId: 0,
		Limit: math.MaxInt,
	}
	ownerKeys, err := c.keyRepository.FetchKeyPairsOfUser(ctx, transactionId, user, keyPagination)
	if err != nil {
		return nil, err
	}

	var ownerKey *model_server.UserKeyPair = nil
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
		sigGraphIngredients = append(sigGraphIngredients, model_server.ToSigGraphAsset(&ingredients[i]))
	}

	sigGraphOwnerKeyPair := &model_sig_graph.UserKeyPair{
		Public:  ownerKey.Public,
		Private: ownerKey.Private,
	}

	asset, err := c.api.CreateAsset(
		ctx,
		materialName,
		unit,
		quantity,
		sigGraphOwnerKeyPair,
		sigGraphIngredients,
		ingredientSecretIds,
		secretIds,
		ingredientSignatures,
	)

	if err != nil {
		return nil, err
	}

	secretParentIds := map[string]model_server.PrivateId{}
	for i := range ingredients {
		thisId := ingredients[i].Id
		thisSecret := ingredientSecretIds[i]

		if thisSecret != "" {
			thisHash, err := c.hashGeneratorService.GenerateHashedId(ctx, string(thisId), thisSecret)
			if err != nil {
				return nil, err
			}

			otherId := asset.Id
			otherSecret := secretIds[i]
			otherHash, err := c.hashGeneratorService.GenerateHashedId(ctx, string(otherId), otherSecret)
			if err != nil {
				return nil, err
			}

			privateId := model_server.PrivateId{
				ThisId:     thisId,
				ThisHash:   thisHash,
				ThisSecret: thisSecret,

				OtherId:     model_server.NodeId(otherId),
				OtherSecret: otherSecret,
				OtherHash:   otherHash,
			}

			secretParentIds[thisHash] = privateId
		}
	}

	namespace := fmt.Sprintf("%d", user.ID)
	modelNode := model_server.FromSigGraphNode(&asset.Node, 0, namespace, secretParentIds, map[string]model_server.PrivateId{})
	modelAsset := model_server.FromSigGraphAsset(asset, &modelNode)

	// save asset
	c.repository.SaveAsset(ctx, transactionId, &modelAsset)

	return &modelAsset, nil
}

func (c *assetController) GetAssetById(ctx context.Context, user *model_server.User, id model_server.NodeId, useCache bool) (*model_server.Asset, error) {
	var transactionId repository_server.TransactionId
	var err error
	if useCache {
		transactionId, err = c.transactionManager.BypassTransaction(ctx)
		if err != nil {
			return nil, err
		}
		defer c.transactionManager.StopBypassedTransaction(ctx, transactionId)

		assets, err := c.repository.FetchAssetsByIds(ctx, transactionId, fmt.Sprintf("%d", user.ID), map[model_server.NodeId]bool{id: true})
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

	privateParentIds := map[string]model_server.PrivateId{}
	for hash := range asset.PrivateParentsHashedIds {
		privateParentIds[hash] = model_server.PrivateId{}
	}

	privateChildrenIds := map[string]model_server.PrivateId{}
	for hash := range asset.PrivateChildrenHashedIds {
		privateChildrenIds[hash] = model_server.PrivateId{}
	}

	modelNode := model_server.FromSigGraphNode(&asset.Node, 0, fmt.Sprintf("%d", user.ID), privateParentIds, privateChildrenIds)
	modelAsset := model_server.FromSigGraphAsset(asset, &modelNode)

	c.repository.SaveAsset(ctx, transactionId, &modelAsset)

	return &modelAsset, nil
}

func (c *assetController) GetOwnedAssetsFromCache(
	ctx context.Context,
	user *model_server.User,
	isTransferred []bool,
	pagination repository_server.PaginationOption[model_server.NodeDbId],
) ([]model_server.Asset, error) {
	transactionId, err := c.transactionManager.BypassTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer c.transactionManager.StopBypassedTransaction(ctx, transactionId)

	keyPagination := repository_server.PaginationOption[model_server.UserKeyPairId]{
		MinId: 0,
		Limit: math.MaxInt,
	}
	keys, err := c.keyRepository.FetchKeyPairsOfUser(ctx, transactionId, user, keyPagination)
	if err != nil {
		return nil, err
	}

	ret := []model_server.Asset{}
	for _, key := range keys {
		remainingLimit := pagination.Limit - len(ret)

		assetPagination := repository_server.PaginationOption[model_server.NodeDbId]{
			MinId: pagination.MinId,
			Limit: remainingLimit,
		}
		assets, err := c.repository.FetchAssetsByOwner(
			ctx,
			transactionId,
			fmt.Sprintf("%d", user.ID),
			key.Public,
			isTransferred,
			assetPagination,
		)
		if err != nil {
			return nil, err
		}
		ret = append(ret, assets...)
	}

	return ret, nil
}

func (c *assetController) GetAssetsFromCacheByDbId(
	ctx context.Context,
	user *model_server.User,
	ids map[model_server.NodeDbId]bool,
) ([]model_server.Asset, error) {
	transactionId, err := c.transactionManager.BypassTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer c.transactionManager.StopBypassedTransaction(ctx, transactionId)

	ret := []model_server.Asset{}
	namespace := fmt.Sprintf("%d", user.ID)
	ret, err = c.repository.FetchAssetsByDbIds(ctx, transactionId, namespace, ids)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
