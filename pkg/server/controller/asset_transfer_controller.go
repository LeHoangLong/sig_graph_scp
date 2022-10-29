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
	hashedIdGenerator       utility.HashedIdGeneratorServiceI
	transferApi             api_asset_transfer.AssetTransferServiceApi
	assetRepository         repository_server.AssetRepositoryI
	transactionManager      repository_server.TransactionManagerI
	keyRepository           repository_server.UserKeyRepositoryI
	nodeRepository          repository_server.NodeRepositoryI
	nodeController          NodeControllerI
	peerRepository          repository_server.PeerRepositoryI
	assetTransferRepository repository_server.AssetTransferRepositoryI
	assetController         AssetControllerI
	bus                     EventBus.Bus
}

func NewAssetTransferController(
	clock utility.ClockI,
	hashedIdGenerator utility.HashedIdGeneratorServiceI,
	transferApi api_asset_transfer.AssetTransferServiceApi,
	assetRepository repository_server.AssetRepositoryI,
	transactionManager repository_server.TransactionManagerI,
	keyRepository repository_server.UserKeyRepositoryI,
	nodeRepository repository_server.NodeRepositoryI,
	nodeController NodeControllerI,
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
		nodeController:          nodeController,
		hashedIdGenerator:       hashedIdGenerator,
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
		"",
		response,
	)

	err = c.assetTransferRepository.CreateAssetAcceptRequest(
		ctx,
		tx,
		&modelResponse,
	)

	if err != nil {
		return nil, err
	}

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

	candidateIds := []model_server.CandidateId{}
	for i := range event.Candidates {
		candidateIds = append(candidateIds, model_server.CandidateId{
			Id:        event.Candidates[i].Id,
			Secret:    event.Candidates[i].Secret,
			Signature: event.Candidates[i].Signature,
		})
	}

	asset, err := c.assetController.GetAssetById(ctx, user, model_server.NodeId(event.AssetId), true)
	if err != nil {
		return
	}

	assetTransferRequest := model_server.RequestToAcceptAsset{
		Status:                    model.ERequestToAcceptAssetStatusPending,
		IsOutboundOrInbound:       false,
		Time:                      event.TimeMs,
		AckId:                     event.AckId,
		AssetId:                   asset.NodeDbId,
		PeerId:                    selectedPeer.PeerDbId,
		UserId:                    user.ID,
		ExposedPrivateConnections: exposedPrivateConnections,
		CandidateIds:              candidateIds,
	}

	err = c.assetTransferRepository.CreateAssetAcceptRequest(
		ctx,
		txId,
		&assetTransferRequest,
	)
	if err != nil {
		return
	}
}

func (c *assetTransferController) GetRequestsToAcceptAsset(
	ctx context.Context,
	user *model_server.User,
	status model.ERequestToAcceptAssetStatus,
	inboundOrOutbound bool,
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
		inboundOrOutbound,
		pagination,
	)

	if err != nil {
		return nil, err
	}

	return requests, nil
}

func (c *assetTransferController) AcceptReceivedRequestsToAcceptAsset(
	ctx context.Context,
	user *model_server.User,
	keyPairId model_server.UserKeyPairId,
	requestId model_server.RequestId,
	acceptOrRejct bool,
	message string,
	isNewConnectionSecretOrPublic bool,
	toInformSenderOfNewId bool,
) (*model_server.RequestToAcceptAsset, error) {
	txId, err := c.transactionManager.BypassTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer c.transactionManager.StopBypassedTransaction(ctx, txId)

	request, err := c.assetTransferRepository.FetchAssetAcceptRequestsById(
		ctx,
		txId,
		user,
		requestId,
	)
	if err != nil {
		return nil, err
	}

	namespace := fmt.Sprintf("%d", user.ID)
	asset, err := c.assetRepository.FetchAssetsByDbIds(
		ctx,
		txId,
		namespace,
		map[model_server.NodeDbId]bool{
			request.AssetId: true,
		},
	)
	if err != nil {
		return nil, err
	}
	if len(asset) == 0 {
		return nil, utility.ErrNotFound
	}
	sigGraphAsset := model_server.ToSigGraphAsset(&asset[0])

	peer, err := c.peerRepository.FetchPeerById(
		ctx,
		txId,
		request.PeerId,
	)
	if err != nil {
		return nil, err
	}
	assetTransferPeer := model_server.ToAssetTransferPeer(peer)

	userKeys, err := c.keyRepository.FetchKeyPairsByIds(
		ctx,
		txId,
		user,
		map[model_server.UserKeyPairId]bool{
			keyPairId: true,
		},
	)
	if err != nil {
		return nil, err
	}
	if len(userKeys) == 0 {
		return nil, utility.ErrNotFound
	}
	userKey := userKeys[0]

	assetTransferRequest := model_server.ToAssetTransferRequestToAcceptAsset(
		&sigGraphAsset,
		peer.PeerPemPublicKey,
		model_server.ToSigGraphUserKeyPair(&userKey),
		request,
	)

	err = c.transferApi.AcceptRequestToAcceptAsset(
		ctx,
		&assetTransferPeer,
		&assetTransferRequest,
		acceptOrRejct,
		message,
		isNewConnectionSecretOrPublic,
		toInformSenderOfNewId,
	)

	if err != nil {
		return nil, err
	}

	// update request
	if acceptOrRejct {
		request.Status = model.ERequestToAcceptAssetStatusAccepted
	} else {
		request.Status = model.ERequestToAcceptAssetStatusRejected
	}
	request.AcceptMessage = message
	err = c.assetTransferRepository.UpdateAssetAcceptRequest(ctx, txId, request)
	if err != nil {
		return nil, err
	}

	return request, nil
}

func (c *assetTransferController) SubscribeNewAssetAcceptReceivedEvent(
	ctx context.Context,
	bus EventBus.Bus,
	topic string,
) error {
	return bus.Subscribe(topic, c.newAcceptAssetReceivedHandler)
}

func (c *assetTransferController) newAcceptAssetReceivedHandler(
	event model_asset_transfer.AcceptAssetEvent,
) {
	// TODO: if this function fails, we lose the secret id information. So we might actually want
	// to store in some backup first.
	ctx := context.Background()

	txId, err := c.transactionManager.BypassTransaction(ctx)
	if err != nil {
		return
	}
	defer c.transactionManager.StopBypassedTransaction(ctx, txId)

	request, err := c.assetTransferRepository.FetchAssetAcceptRequestsByAckId(
		ctx,
		txId,
		event.AckId,
		true,
	)

	if err != nil {
		return
	}

	// update request status
	{
		if event.IsAccepted {
			request.Status = model.ERequestToAcceptAssetStatusAccepted
		} else {
			request.Status = model.ERequestToAcceptAssetStatusRejected
		}
		request.AcceptMessage = event.Message
		c.assetTransferRepository.UpdateAssetAcceptRequest(
			ctx,
			txId,
			request,
		)
	}

	user := model_server.User{
		ID: request.UserId,
	}

	if event.NewId != "" {
		// fetch transferred asset from sig graph
		c.assetController.GetAssetById(
			ctx,
			&user,
			model_server.NodeId(event.NewId),
			false,
		)
	}

	// update cached current asset
	{
		namespace := fmt.Sprintf("%d", user.ID)
		currentAssets, err := c.assetRepository.FetchAssetsByDbIds(
			ctx,
			txId,
			namespace,
			map[model_server.NodeDbId]bool{
				request.AssetId: true,
			},
		)
		if err != nil {
			return
		}

		currentAsset := &currentAssets[0]
		// refetch from sig graph
		currentAsset, err = c.assetController.GetAssetById(
			ctx,
			&user,
			currentAsset.Id,
			false,
		)
		if err != nil {
			return
		}

		if event.NewId == "" || event.NewSecret == "" {
			return
		}

		if currentAsset.Id != model_server.NodeId(event.OldId) {
			// hmm, something wrong
		}

		// update current asset's private edge toward the new asset
		newHash, err := c.hashedIdGenerator.GenerateHashedId(event.NewId, event.NewSecret)
		if err != nil {
			return
		}

		oldHash, err := c.hashedIdGenerator.GenerateHashedId(event.OldId, event.OldSecret)
		if err != nil {
			return
		}

		fmt.Println("Updating node secret id")
		c.nodeController.UpdateNodeSecretId(
			ctx,
			&currentAsset.Node,
			map[string]model_server.PrivateId{
				newHash: {
					ThisId:      model_server.NodeId(event.NewId),
					ThisSecret:  event.NewSecret,
					ThisHash:    newHash,
					OtherId:     currentAsset.Id,
					OtherSecret: event.OldSecret,
					OtherHash:   oldHash,
				},
			},
		)
	}
}
