package controller_server

import (
	"context"
	"fmt"
	"math"
	api_asset_transfer "sig_graph_scp/pkg/asset_transfer/api"
	model_asset_transfer "sig_graph_scp/pkg/asset_transfer/model"
	"sig_graph_scp/pkg/model"
	model_server "sig_graph_scp/pkg/server/model"
	repository_server "sig_graph_scp/pkg/server/repository"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"
	"sig_graph_scp/pkg/utility"

	EventBus "github.com/asaskevich/eventbus"
)

type assetTransferController struct {
	clock                   utility.ClockI
	transferApi             api_asset_transfer.AssetTransferServiceApi
	assetRepository         repository_server.AssetRepositoryI
	transactionManager      repository_server.TransactionManagerI
	keyRepository           repository_server.UserKeyRepositoryI
	nodeRepository          repository_server.NodeRepositoryI
	peerRepository          repository_server.PeerRepositoryI
	assetTransferRepository repository_server.AssetTransferRepositoryI
	assetController         AssetControllerI
	bus                     EventBus.Bus
}

func NewAssetTransferController(
	clock utility.ClockI,
	transferApi api_asset_transfer.AssetTransferServiceApi,
	assetRepository repository_server.AssetRepositoryI,
	transactionManager repository_server.TransactionManagerI,
	keyRepository repository_server.UserKeyRepositoryI,
	nodeRepository repository_server.NodeRepositoryI,
	peerRepository repository_server.PeerRepositoryI,
	assetController AssetControllerI,
	assetTransferRepository repository_server.AssetTransferRepositoryI,
) *assetTransferController {
	return &assetTransferController{
		clock:                   clock,
		transferApi:             transferApi,
		assetRepository:         assetRepository,
		transactionManager:      transactionManager,
		keyRepository:           keyRepository,
		nodeRepository:          nodeRepository,
		peerRepository:          peerRepository,
		assetController:         assetController,
		assetTransferRepository: assetTransferRepository,
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
	keyPagination := repository_server.PaginationOption[model_server.UserKeyPairId]{
		MinId: 0,
		Limit: math.MaxInt,
	}
	keys, err := c.keyRepository.FetchKeyPairsOfUser(ctx, tx, user, keyPagination)
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

func (c *assetTransferController) SubscribeNewAcceptAssetRequestReceivedEvent(
	ctx context.Context,
	bus EventBus.Bus,
	topic string,
) error {
	return bus.Subscribe(topic, c.newAcceptAssetRequestReceivedHandler)
}

// TODO: add callback to inform of result for handling or inform sender
// that it has failed and they need to retry
func (c *assetTransferController) newAcceptAssetRequestReceivedHandler(
	event model_asset_transfer.RequestToAcceptAssetEvent,
) {
	ctx := context.Background()

	txId, err := c.transactionManager.BypassTransaction(ctx)
	if err != nil {
		return
	}
	defer c.transactionManager.StopBypassedTransaction(ctx, txId)

	user, err := c.keyRepository.FetchUserWithPublicKey(ctx, txId, event.UserPemPublicKey)
	if err != nil {
		return
	}

	peerPagination := repository_server.PaginationOption[model_server.PeerDbId]{
		MinId: 0,
		Limit: math.MaxInt,
	}
	peers, err := c.peerRepository.FetchPeersByUser(ctx, txId, user, peerPagination)
	if err != nil {
		return
	}
	var selectedPeer *model_server.Peer
	for i := range peers {
		if peers[i].PeerPemPublicKey == event.PeerPemPublicKey {
			selectedPeer = &peers[i]
			break
		}
	}
	if selectedPeer == nil {
		return
	}

	exposedPrivateConnections := map[string]model_server.PrivateId{}
	for hash := range event.ExposedPrivateConnections {
		exposedPrivateConnections[hash] = model_server.PrivateId{
			ThisId:     model_server.NodeId(event.ExposedPrivateConnections[hash].ThisId),
			ThisSecret: event.ExposedPrivateConnections[hash].ThisSecret,
			ThisHash:   event.ExposedPrivateConnections[hash].ThisHash,

			OtherId:     model_server.NodeId(event.ExposedPrivateConnections[hash].OtherId),
			OtherSecret: event.ExposedPrivateConnections[hash].OtherSecret,
			OtherHash:   event.ExposedPrivateConnections[hash].OtherHash,
		}
	}

	fmt.Println("event.AssetId ", event.AssetId)
	asset, err := c.assetController.GetAssetById(ctx, user, model_server.NodeId(event.AssetId), true)
	if err != nil {
		return
	}

	assetTransferRequest := model_server.RequestToAcceptAsset{
		Status:                    model.ERequestToAcceptAssetStatusPending,
		IsOutboundOrInbound:       false,
		Time:                      event.TimeMs,
		AckId:                     event.AckId,
		Accepted:                  false,
		AssetId:                   asset.NodeDbId,
		PeerId:                    selectedPeer.PeerDbId,
		UserId:                    user.ID,
		ExposedPrivateConnections: exposedPrivateConnections,
	}

	err = c.assetTransferRepository.CreateAssetAcceptRequest(
		ctx,
		txId,
		&assetTransferRequest,
	)
	if err != nil {
		return
	}

	fmt.Printf("saved %+v\n", assetTransferRequest)
}

func (c *assetTransferController) GetReceivedRequestsToAcceptAsset(
	ctx context.Context,
	user *model_server.User,
	status model.ERequestToAcceptAssetStatus,
	pagination repository_server.PaginationOption[model_server.RequestId],
) ([]model_server.RequestToAcceptAsset, error) {
	txId, err := c.transactionManager.BypassTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer c.transactionManager.StopBypassedTransaction(ctx, txId)

	requests, err := c.assetTransferRepository.FetchAssetAcceptRequestsByUserAndStatus(
		ctx,
		txId,
		user,
		status,
		pagination,
	)
	if err != nil {
		return nil, err
	}

	return requests, nil
}
