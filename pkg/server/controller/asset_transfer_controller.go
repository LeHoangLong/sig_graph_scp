package controller_server

import (
	"context"
	"fmt"
	api_asset_transfer "sig_graph_scp/pkg/asset_transfer/api"
	model_asset_transfer "sig_graph_scp/pkg/asset_transfer/model"
	model_server "sig_graph_scp/pkg/server/model"
	repository_server "sig_graph_scp/pkg/server/repository"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"
	"sig_graph_scp/pkg/utility"
)

type assetTransferController struct {
	clock              utility.ClockI
	transferApi        api_asset_transfer.AssetTransferServiceApi
	assetRepository    repository_server.AssetRepositoryI
	transactionManager repository_server.TransactionManagerI
	keyRepository      repository_server.UserKeyRepositoryI
	nodeRepository     repository_server.NodeRepositoryI
	peerRepository     repository_server.PeerRepositoryI
}

func NewAssetTransferController(
	clock utility.ClockI,
	transferApi api_asset_transfer.AssetTransferServiceApi,
	assetRepository repository_server.AssetRepositoryI,
	transactionManager repository_server.TransactionManagerI,
	keyRepository repository_server.UserKeyRepositoryI,
	nodeRepository repository_server.NodeRepositoryI,
	peerRepository repository_server.PeerRepositoryI,
) *assetTransferController {
	return &assetTransferController{
		clock:              clock,
		transferApi:        transferApi,
		assetRepository:    assetRepository,
		transactionManager: transactionManager,
		keyRepository:      keyRepository,
		nodeRepository:     nodeRepository,
		peerRepository:     peerRepository,
	}
}

func (c *assetTransferController) TransferAsset(
	ctx context.Context,
	user *model_server.User,
	assetId model_server.NodeDbId,
	peerId model_server.PeerDbId,
	exposedSecretIds []repository_server.EdgeNodeId,
	isNewConnectionPublicOrPrivate bool,
) (*model_server.RequestToAcceptAsset, error) {
	time := c.clock.Now()

	tx, err := c.transactionManager.BypassTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer c.transactionManager.StopBypassedTransaction(ctx, tx)

	namespace := fmt.Sprintf("%d", user.ID)

	// fetch peer
	peer, err := c.peerRepository.FetchPeerById(ctx, tx, peerId)
	if err != nil {
		return nil, err
	}

	// fetch assets
	assets, err := c.assetRepository.FetchAssetsByDbIds(
		ctx,
		tx,
		namespace,
		map[model_server.NodeDbId]bool{assetId: true},
	)
	if err != nil {
		return nil, err
	}

	if len(assets) == 0 {
		return nil, fmt.Errorf("%w: no such asset id", utility.ErrNotFound)
	}

	asset := &assets[0]

	sigGraphAsset := model_server.ToSigGraphAsset(asset)

	// find keys
	keys, err := c.keyRepository.FetchKeyPairsOfUser(ctx, tx, user)
	if err != nil {
		return nil, err
	}
	var selectedKey *model_server.UserKeyPair
	for i := range keys {
		if keys[i].Public == asset.OwnerPublicKey {
			selectedKey = &keys[i]
			break
		}
	}

	if selectedKey == nil {
		return nil, fmt.Errorf("%w: could not find public key %s", utility.ErrNotFound, asset.OwnerPublicKey)
	}

	secretIds, err := c.nodeRepository.FetchPrivateEdgesByNodeIds(
		ctx,
		tx,
		namespace,
		exposedSecretIds,
	)
	if err != nil {
		return nil, err
	}

	assetTransferPrivateIds := map[string]model_asset_transfer.PrivateId{}
	for i := range secretIds {
		assetTransferPrivateIds[secretIds[i].ThisHash] = model_server.ToAssetTransferPrivateId(&secretIds[i])
	}

	sigGraphKey := model_sig_graph.UserKeyPair{
		Public:  selectedKey.Public,
		Private: selectedKey.Private,
	}
	assetTransferPeer := model_server.ToAssetTransferPeer(peer)

	response, err := c.transferApi.TransferAsset(
		ctx,
		time,
		&sigGraphAsset,
		&sigGraphKey,
		&assetTransferPeer,
		assetTransferPrivateIds,
		isNewConnectionPublicOrPrivate,
	)
	if err != nil {
		return nil, err
	}

	modelResponse := model_server.FromAssetTransferRequestToAcceptAsset(
		0,
		assetId,
		peer.PeerDbId,
		user.ID,
		response,
	)
	return &modelResponse, nil
}
