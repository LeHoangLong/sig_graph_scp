package api_asset_transfer

import (
	"context"
	service_asset_transfer "sig_graph_scp/internal/asset_transfer/service"
	service_sig_graph "sig_graph_scp/internal/sig_graph/service"
	model_asset_transfer "sig_graph_scp/pkg/asset_transfer/model"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"
	"sig_graph_scp/pkg/utility"
	"time"
)

type AssetTransferServiceApi interface {
	TransferAsset(
		ctx context.Context,
		requestTime time.Time,
		asset *model_sig_graph.Asset,
		ownerKey *model_sig_graph.UserKeyPair,
		peer *model_asset_transfer.Peer,
		exposedPrivateConnections map[string]model_asset_transfer.PrivateId,
		isNewConnectionSecretOrPublic bool,
	) (*model_asset_transfer.RequestToAcceptAsset, error)

	AcceptRequestToAcceptAsset(
		ctx context.Context,
		peer *model_asset_transfer.Peer,
		request *model_asset_transfer.RequestToAcceptAsset,
		acceptOrReject bool,
		message string,
		isNewConnectionSecretOrPublic bool,
		toInformSenderOfNewId bool,
	) error
}

type Options struct {
	// number of candidate id to generate when transfer asset
	NumberOfCandidates uint32
}

type assetTransferServiceApi struct {
	assetTransferService service_asset_transfer.AssetTransferServiceI
}

func NewAssetTransferServiceApi(graphName string, options *Options) (AssetTransferServiceApi, error) {
	connPool := utility.NewGrpcConnectionPool()
	numberOfCandidates := uint32(6)
	if options != nil {
		numberOfCandidates = options.NumberOfCandidates
	}
	secretGenerator := service_asset_transfer.NewSecretIdGeneratorCrypto(20)
	idGenerator := service_sig_graph.NewIdGenerateServiceUuid(graphName)
	nodeSigningService := service_sig_graph.NewNodeSigningService()

	assetTransferService := service_asset_transfer.NewAssetTransferServiceGrpc(
		connPool,
		numberOfCandidates,
		secretGenerator,
		idGenerator,
		nodeSigningService,
	)

	return &assetTransferServiceApi{
		assetTransferService: assetTransferService,
	}, nil
}

func (s *assetTransferServiceApi) TransferAsset(
	ctx context.Context,
	requestTime time.Time,
	asset *model_sig_graph.Asset,
	ownerKey *model_sig_graph.UserKeyPair,
	peer *model_asset_transfer.Peer,
	exposedPrivateConnections map[string]model_asset_transfer.PrivateId,
	isNewConnectionSecretOrPublic bool,
) (*model_asset_transfer.RequestToAcceptAsset, error) {
	return s.assetTransferService.TransferAsset(
		ctx,
		requestTime,
		asset,
		ownerKey,
		peer,
		exposedPrivateConnections,
		isNewConnectionSecretOrPublic,
	)
}

func (s *assetTransferServiceApi) AcceptRequestToAcceptAsset(
	ctx context.Context,
	peer *model_asset_transfer.Peer,
	request *model_asset_transfer.RequestToAcceptAsset,
	acceptOrReject bool,
	message string,
	isNewConnectionSecretOrPublic bool,
	toInformSenderOfNewId bool,
) error {
	return s.assetTransferService.AcceptRequestToAcceptAsset(
		ctx,
		peer,
		request,
		acceptOrReject,
		message,
		isNewConnectionSecretOrPublic,
		toInformSenderOfNewId,
	)
}
