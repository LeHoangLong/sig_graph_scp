package service_asset_transfer

import (
	"context"
	"crypto/sha512"
	"fmt"
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
	connPool           utility.GrpcConnectionPoolI
	numberOfCandidate  uint32
	secretIdGeneratorI SecretIdGeneratorI
	idGeneratorService service_sig_graph.IdGenerateServiceI
	nodeSigningService service_sig_graph.NodeSigningServiceI
	sigGraphClientApi  api_sig_graph.SigGraphClientApi
}

func NewAssetTransferServiceGrpc(
	connPool utility.GrpcConnectionPoolI,
	numberOfCandidate uint32,
	secretIdGeneratorI SecretIdGeneratorI,
	idGeneratorService service_sig_graph.IdGenerateServiceI,
	nodeSigningService service_sig_graph.NodeSigningServiceI,
) *assetTransferServiceGrpc {
	return &assetTransferServiceGrpc{
		connPool:           connPool,
		numberOfCandidate:  numberOfCandidate,
		secretIdGeneratorI: secretIdGeneratorI,
		idGeneratorService: idGeneratorService,
		nodeSigningService: nodeSigningService,
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

	if isNewConnectionSecretOrPublic {
		for i := uint32(0); i < s.numberOfCandidate; i++ {
			id, err := s.idGeneratorService.NewFullId(ctx)
			if err != nil {
				return nil, err
			}

			secret, err := s.secretIdGeneratorI.NewSecretId(ctx)
			if err != nil {
				return nil, err
			}

			// generate signature for this candidate
			hashByte := sha512.Sum512([]byte(fmt.Sprintf("%s%s", id, secret)))
			hash := string(hashByte[:])
			asset.PrivateChildrenHashedIds[hash] = true
			signature, err := s.nodeSigningService.Sign(ctx, ownerKey, &asset)
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
	} else {
		for i := uint32(0); i < s.numberOfCandidate; i++ {
			id, err := s.idGeneratorService.NewFullId(ctx)
			if err != nil {
				return nil, err
			}

			// generate signature for this candidate
			asset.PublicChildrenIds[string(id)] = true
			signature, err := s.nodeSigningService.Sign(ctx, ownerKey, &asset)
			if err != nil {
				return nil, err
			}

			newCandidate := sig_graph_grpc.SignatureCandidate{
				Id:        string(id),
				Secret:    "",
				Signature: signature,
			}
			candidates = append(candidates, &newCandidate)
		}
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
	toInformSenderOfNewId bool,
) error {
	conn, err := s.connPool.NewConnection(ctx, peer.ConnectionUri)
	if err != nil {
		return err
	}
	defer s.connPool.ReturnConnection(ctx, peer.ConnectionUri, conn)

	grpcRequest := sig_graph_grpc.AcceptAssetRequest{
		AckId:     request.AckId,
		Accepted:  acceptOrReject,
		Message:   message,
		NewId:     "",
		NewSecret: "",
		OldId:     "",
		OldSecret: "",
	}

	if toInformSenderOfNewId {
		grpcRequest.NewId = "test-id"
		grpcRequest.NewSecret = "test-secert"
		grpcRequest.OldId = "old-id"
		grpcRequest.OldSecret = "old-secret"
	}

	client := sig_graph_grpc.NewTransferAssetClient(conn)
	_, err = client.AcceptAsset(ctx, &grpcRequest)
	if err != nil {
		return err
	}

	return nil
}
