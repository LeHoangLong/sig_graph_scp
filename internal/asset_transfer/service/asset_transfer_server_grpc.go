package service_asset_transfer

import (
	"context"
	"net"
	utility_asset_transfer "sig_graph_scp/internal/asset_transfer/utility"
	sig_graph_grpc "sig_graph_scp/internal/grpc"
	model_asset_transfer "sig_graph_scp/pkg/asset_transfer/model"
	"sig_graph_scp/pkg/utility"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type assetTransferServerGrpc struct {
	sig_graph_grpc.UnimplementedTransferAssetServer
	mtx                    utility.MutexI
	requestToAcceptHandler AssetTransferHandlerI
	assetAcceptHandler     AssetAcceptHandlerI
	address                string
	hashGenerator          utility.HashedIdGeneratorServiceI
}

func NewAssetTransferServerGrpc(
	requestToAcceptHandler AssetTransferHandlerI,
	assetAcceptHandler AssetAcceptHandlerI,
	address string,
	hashGenerator utility.HashedIdGeneratorServiceI,
) *assetTransferServerGrpc {
	return &assetTransferServerGrpc{
		mtx:                    utility.NewMutex(),
		requestToAcceptHandler: requestToAcceptHandler,
		assetAcceptHandler:     assetAcceptHandler,
		address:                address,
		hashGenerator:          hashGenerator,
	}
}

func (s *assetTransferServerGrpc) RegisterHandler(
	ctx context.Context,
	requestToAcceptHandler AssetTransferHandlerI,
) error {
	if !s.mtx.Lock(ctx) {
		return utility.ErrTimedOut
	}

	defer s.mtx.Unlock(ctx)
	s.requestToAcceptHandler = requestToAcceptHandler
	return nil
}

func (s *assetTransferServerGrpc) RegisterAssetAcceptHandler(
	ctx context.Context,
	requestToAcceptHandler AssetAcceptHandlerI,
) error {
	s.assetAcceptHandler = requestToAcceptHandler
	return nil
}

func (s *assetTransferServerGrpc) Start() error {
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()

	sig_graph_grpc.RegisterTransferAssetServer(grpcServer, s)

	grpcServer.Serve(lis)
	return nil
}

func (s *assetTransferServerGrpc) RequestToAcceptAsset(
	ctx context.Context,
	request *sig_graph_grpc.RequestToAcceptAssetRequest,
) (*sig_graph_grpc.RequestToAcceptAssetResponse, error) {
	if !s.mtx.Lock(ctx) {
		return &sig_graph_grpc.RequestToAcceptAssetResponse{
			Error: utility_asset_transfer.ToGrpcError(utility.ErrTimedOut),
		}, nil
	}

	handler := s.requestToAcceptHandler
	s.mtx.Unlock(ctx)

	requestTime := time.UnixMilli(int64(request.TimeMs))
	assetId := request.AssetId
	senderPublicKey := request.OwnerPublicKey
	recipientPublicKey := request.NewOwnerPublicKey

	grpcExposedSecretIds := request.SecretIds
	exposedSecretIds := map[string]model_asset_transfer.PrivateId{}
	for hash, id := range grpcExposedSecretIds {
		thisHash, err := s.hashGenerator.GenerateHashedId(ctx, id.ThisId, id.ThisSecret)
		if err != nil {
			return &sig_graph_grpc.RequestToAcceptAssetResponse{
				Error: utility_asset_transfer.ToGrpcError(err),
			}, nil
		}

		otherHash, err := s.hashGenerator.GenerateHashedId(ctx, id.OtherId, id.OtherSecret)
		if err != nil {
			return &sig_graph_grpc.RequestToAcceptAssetResponse{
				Error: utility_asset_transfer.ToGrpcError(err),
			}, nil
		}

		exposedSecretIds[hash] = model_asset_transfer.PrivateId{
			ThisId:     id.ThisId,
			ThisSecret: id.ThisSecret,
			ThisHash:   thisHash,

			OtherId:     id.OtherId,
			OtherSecret: id.OtherSecret,
			OtherHash:   otherHash,
		}
	}

	candidates := []model_asset_transfer.CandidateId{}
	for i := range request.Candidates {
		candidates = append(candidates, model_asset_transfer.CandidateId{
			Id:        request.Candidates[i].Id,
			Secret:    request.Candidates[i].Secret,
			Signature: request.Candidates[i].Signature,
		})
	}

	ackId := uuid.New().String()
	err := handler.HandleAssetTransfer(ctx, ackId, &requestTime, assetId, senderPublicKey, recipientPublicKey, exposedSecretIds, candidates)
	if err != nil {
		return &sig_graph_grpc.RequestToAcceptAssetResponse{
			Error: utility_asset_transfer.ToGrpcError(err),
		}, nil
	}

	return &sig_graph_grpc.RequestToAcceptAssetResponse{
		AckId: ackId,
	}, nil
}

func (s *assetTransferServerGrpc) AcceptAsset(
	ctx context.Context,
	request *sig_graph_grpc.AcceptAssetRequest,
) (*sig_graph_grpc.AcceptAssetResponse, error) {
	if !s.mtx.Lock(ctx) {
		return &sig_graph_grpc.AcceptAssetResponse{}, nil
	}

	handler := s.assetAcceptHandler
	s.mtx.Unlock(ctx)

	handler.HandleAssetAccept(
		ctx,
		request.AckId,
		request.Accepted,
		request.Message,
		request.NewId,
		request.NewSecret,
		request.OldId,
		request.OldSecret,
	)

	return &sig_graph_grpc.AcceptAssetResponse{}, nil
}
