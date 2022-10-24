package service_asset_transfer

import (
	"context"
	"crypto/sha512"
	"fmt"
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
	mtx     utility.MutexI
	handler AssetTransferHandlerI
	address string
}

func NewAssetTransferServerGrpc(
	handler AssetTransferHandlerI,
	address string,
) *assetTransferServerGrpc {
	return &assetTransferServerGrpc{
		mtx:     utility.NewMutex(),
		handler: handler,
		address: address,
	}
}

func (s *assetTransferServerGrpc) RegisterHandler(
	ctx context.Context,
	handler AssetTransferHandlerI,
) error {
	if !s.mtx.Lock(ctx) {
		return utility.ErrTimedOut
	}

	defer s.mtx.Unlock(ctx)
	s.handler = handler
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

	handler := s.handler
	s.mtx.Unlock(ctx)

	requestTime := time.UnixMilli(int64(request.TimeMs))
	assetId := request.AssetId
	senderPublicKey := request.OwnerPublicKey
	recipientPublicKey := request.NewOwnerPublicKey

	grpcExposedSecretIds := request.SecretIds
	exposedSecretIds := map[string]model_asset_transfer.PrivateId{}
	for hash, id := range grpcExposedSecretIds {
		thisSecretId := fmt.Sprintf("%s%s", id.ThisId, id.ThisSecret)
		thisHashBytes := sha512.Sum512([]byte(thisSecretId))
		thisHash := string(thisHashBytes[:])

		otherSecretId := fmt.Sprintf("%s%s", id.OtherId, id.OtherSecret)
		otherHashBytes := sha512.Sum512([]byte(otherSecretId))
		otherHash := string(otherHashBytes[:])

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
	panic("implement me")
}
