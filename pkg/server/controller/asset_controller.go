package controller_server

import (
	"context"
	"fmt"
	service_sig_graph "sig_graph_scp/internal/sig_graph/service"
	"sig_graph_scp/pkg/model"
	repository_server "sig_graph_scp/pkg/server/repository"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"
)

type assetController struct {
	service            service_sig_graph.AssetServiceI
	repository         repository_server.AssetRepositoryI
	keyRepository      repository_server.UserKeyRepositoryI
	transactionManager repository_server.TransactionManagerI
}

func NewAssetController(
	service service_sig_graph.AssetServiceI,
	repository repository_server.AssetRepositoryI,
	keyRepository repository_server.UserKeyRepositoryI,
	transactionManager repository_server.TransactionManagerI,
) AssetControllerI {
	return &assetController{
		service:            service,
		repository:         repository,
		keyRepository:      keyRepository,
		transactionManager: transactionManager,
	}
}

func (c *assetController) GetAssetById(ctx context.Context, user *model.User, id model.NodeId, useCache bool) (*model.Asset, error) {
	var transactionId repository_server.TransactionId
	var err error
	if useCache {
		transactionId, err = c.transactionManager.BypassTransaction(ctx)
		if err != nil {
			return nil, err
		}

		assets, err := c.repository.FetchAssetsByIds(ctx, transactionId, fmt.Sprintf("%d", user.ID), map[model.NodeId]bool{id: true})
		if err != nil {
			return nil, err
		}

		if len(assets) > 0 {
			return &assets[0], nil
		}
	}

	asset, err := c.service.GetAssetById(ctx, id)
	if err != nil {
		return nil, err
	}

	modelNode := model_sig_graph.ToModelNode(&asset.Node, 0, fmt.Sprintf("%d", user.ID), map[model.PrivateId]bool{}, map[model.PrivateId]bool{})
	modelAsset := model_sig_graph.ToModelAsset(asset, &modelNode)

	if useCache {
		c.repository.SaveAsset(ctx, transactionId, &modelAsset)
	}

	return &modelAsset, nil
}

func (c *assetController) GetOwnedAssetsFromCache(ctx context.Context, user *model.User) ([]model.Asset, error) {
	transactionId, err := c.transactionManager.BypassTransaction(ctx)
	if err != nil {
		return nil, err
	}

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
