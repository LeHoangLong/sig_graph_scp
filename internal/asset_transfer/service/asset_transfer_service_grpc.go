package service_asset_transfer

import (
	"context"
	utility_asset_transfer "sig_graph_scp/internal/asset_transfer/utility"
	sig_graph_grpc "sig_graph_scp/internal/grpc"
	service_sig_graph "sig_graph_scp/internal/sig_graph/service"
	model_asset_transfer "sig_graph_scp/pkg/asset_transfer/model"
	"sig_graph_scp/pkg/model"
	api_sig_graph "sig_graph_scp/pkg/sig_graph/api"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"
	"sig_graph_scp/pkg/utility"
	"time"
)

type assetTransferServiceGrpc struct {
	connPool             utility.GrpcConnectionPoolI
	numberOfCandidate    uint32
	secretIdGeneratorI   SecretIdGeneratorI
	idGeneratorService   service_sig_graph.IdGenerateServiceI
	nodeSigningService   service_sig_graph.NodeSigningServiceI
	sigGraphClientApi    api_sig_graph.SigGraphClientApi
	hashGeneratorService utility.HashedIdGeneratorServiceI
	cloner               utility.ClonerI
}

func NewAssetTransferServiceGrpc(
	connPool utility.GrpcConnectionPoolI,
	numberOfCandidate uint32,
	secretIdGeneratorI SecretIdGeneratorI,
	idGeneratorService service_sig_graph.IdGenerateServiceI,
	nodeSigningService service_sig_graph.NodeSigningServiceI,
	sigGraphClientApi api_sig_graph.SigGraphClientApi,
	hashGeneratorService utility.HashedIdGeneratorServiceI,
	cloner utility.ClonerI,
) *assetTransferServiceGrpc {
	return &assetTransferServiceGrpc{
		connPool:             connPool,
		numberOfCandidate:    numberOfCandidate,
		secretIdGeneratorI:   secretIdGeneratorI,
		idGeneratorService:   idGeneratorService,
		nodeSigningService:   nodeSigningService,
		sigGraphClientApi:    sigGraphClientApi,
		hashGeneratorService: hashGeneratorService,
		cloner:               cloner,
	}
}

func (s *assetTransferServiceGrpc) SetNumberOfCandidatesSignature(
	ctx context.Context,
	numberOfCandidate uint32,
) error {
	s.numberOfCandidate = numberOfCandidate
	return nil
}

func (s *assetTransferServiceGrpc) TransferAsset(
	ctx context.Context,
	requestTime time.Time,
	asset *model_sig_graph.Asset,
	ownerKey *model_sig_graph.UserKeyPair,
	peer *model_asset_transfer.Peer,
	exposedPrivateConnections map[string]model_asset_transfer.PrivateId,
	isNewConnectionSecretOrPublic bool,
) (*model_asset_transfer.RequestToAcceptAsset, error) {
	if peer.Protocol.Type != model.EPeerProtocolGrpc {
		return nil, utility.ErrInvalidArgument
	}

	conn, err := s.connPool.NewConnection(ctx, peer.ConnectionUri)
	if err != nil {
		return nil, err
	}
	defer s.connPool.ReturnConnection(ctx, peer.ConnectionUri, conn)

	secretIds := map[string]*sig_graph_grpc.SecretId{}
	candidates := []*sig_graph_grpc.SignatureCandidate{}
	client := sig_graph_grpc.NewTransferAssetClient(conn)

	for hash := range exposedPrivateConnections {
		secretIds[hash] = &sig_graph_grpc.SecretId{
			ThisId:     string(exposedPrivateConnections[hash].ThisId),
			ThisSecret: exposedPrivateConnections[hash].ThisSecret,

			OtherId:     string(exposedPrivateConnections[hash].OtherId),
			OtherSecret: exposedPrivateConnections[hash].OtherSecret,
		}
	}

	for i := uint32(0); i < s.numberOfCandidate; i++ {
		// generate signature for this candidate
		secret := ""
		signature := ""
		draftAsset := &model_sig_graph.Asset{}

		err = s.cloner.Clone(ctx, asset, draftAsset)
		if err != nil {
			return nil, err
		}
		draftAsset.UpdatedTime = uint64(requestTime.UnixMilli())
		draftAsset.IsFinalized = true

		id, err := s.idGeneratorService.NewFullId(ctx)
		if err != nil {
			return nil, err
		}

		if isNewConnectionSecretOrPublic {
			secret, err = s.secretIdGeneratorI.NewSecretId(ctx)
			if err != nil {
				return nil, err
			}

			hash, err := s.hashGeneratorService.GenerateHashedId(ctx, id, secret)
			if err != nil {
				return nil, err
			}

			draftAsset.PrivateChildrenHashedIds[hash] = true
		} else {
			draftAsset.PublicChildrenIds[string(id)] = true
		}

		signature, err = s.nodeSigningService.Sign(ctx, ownerKey, &draftAsset)
		if err != nil {
			return nil, err
		}
		newCandidate := sig_graph_grpc.SignatureCandidate{
			Id:        string(id),
			Secret:    secret,
			Signature: signature,
		}
		candidates = append(candidates, &newCandidate)
	}

	grpcRequest := sig_graph_grpc.RequestToAcceptAssetRequest{
		TimeMs:            uint64(requestTime.UnixMilli()),
		AssetId:           string(asset.Node.Id),
		OwnerPublicKey:    ownerKey.Public,
		NewOwnerPublicKey: peer.PeerPemPublicKey,
		SecretIds:         secretIds,
		Candidates:        candidates,
	}

	response, err := client.RequestToAcceptAsset(ctx, &grpcRequest)
	if err != nil {
		return nil, err
	}

	err = utility_asset_transfer.WrapGrpcError(response.GetError())
	if err != nil {
		return nil, err
	}

	{
		tempCandidates := []model_asset_transfer.CandidateId{}
		for i := range candidates {
			tempCandidates = append(tempCandidates, model_asset_transfer.CandidateId{
				Id:        candidates[i].Id,
				Secret:    candidates[i].Secret,
				Signature: candidates[i].Signature,
			})
		}

		modelRequest := model_asset_transfer.RequestToAcceptAsset{
			Status:                    model.ERequestToAcceptAssetStatusPending,
			IsOutboundOrInbound:       true,
			TimeMs:                    uint64(requestTime.UnixMilli()),
			AckId:                     response.AckId,
			Accepted:                  false,
			Asset:                     *asset,
			PeerPemPublicKey:          peer.PeerPemPublicKey,
			UserKeyPair:               *ownerKey,
			ExposedPrivateConnections: exposedPrivateConnections,
			Candidates:                tempCandidates,
		}

		return &modelRequest, err
	}
}

func (s *assetTransferServiceGrpc) AcceptRequestToAcceptAsset(
	ctx context.Context,
	peer *model_asset_transfer.Peer,
	request *model_asset_transfer.RequestToAcceptAsset,
	acceptOrReject bool,
	message string,
	isNewConnectionSecretOrPublic bool,
) (updatedRequest *model_asset_transfer.RequestToAcceptAsset, newSecret string, oldSecret string, err error) {
	conn, err := s.connPool.NewConnection(ctx, peer.ConnectionUri)
	if err != nil {
		return
	}
	defer s.connPool.ReturnConnection(ctx, peer.ConnectionUri, conn)

	updatedRequest = &model_asset_transfer.RequestToAcceptAsset{}
	*updatedRequest = *request

	grpcRequest := sig_graph_grpc.AcceptAssetRequest{
		AckId:    updatedRequest.AckId,
		Accepted: acceptOrReject,
		Message:  message,
	}

	if acceptOrReject {
		newSecret, oldSecret, err = s.transferAssetOnSigraphAndUpdateAssetOfRequest(
			ctx,
			isNewConnectionSecretOrPublic,
			updatedRequest,
		)
		if err != nil {
			return
		}
	}

	client := sig_graph_grpc.NewTransferAssetClient(conn)
	_, err = client.AcceptAsset(ctx, &grpcRequest)
	if err != nil {
		return
	}

	return
}

func (s *assetTransferServiceGrpc) transferAssetOnSigraphAndUpdateAssetOfRequest(
	ctx context.Context,
	isNewConnectionSecretOrPublic bool,
	request *model_asset_transfer.RequestToAcceptAsset,
) (newSecret string, oldSecret string, err error) {
	currentSecret := ""
	if isNewConnectionSecretOrPublic {
		currentSecret, err = s.secretIdGeneratorI.NewSecretId(ctx)
		if err != nil {
			return
		}
	}

	var selectedCandidate *model_asset_transfer.CandidateId
	for i := range request.Candidates {
		var newAsset, updatedAsset *model_sig_graph.Asset
		updatedAsset, newAsset, err = s.sigGraphClientApi.TransferAsset(
			ctx,
			request.TimeMs,
			&request.Asset,
			&request.UserKeyPair,
			request.Candidates[i].Id,
			request.Candidates[i].Secret,
			currentSecret,
			request.Candidates[i].Signature,
		)

		if err != nil {
			// if already exists, use another candidate
			if err == utility.ErrAlreadyExists {
				continue
			}
			return
		}

		selectedCandidate = &request.Candidates[i]
		request.NewAsset = newAsset
		request.Asset = *updatedAsset
		break
	}
	newSecret = selectedCandidate.Secret
	oldSecret = currentSecret
	return
}
