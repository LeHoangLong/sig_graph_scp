package service_asset_transfer

import (
	"context"
	model_asset_transfer "sig_graph_scp/pkg/asset_transfer/model"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"
	"time"
)

type AssetTransferServiceI interface {
	// if isNewConnectionSecretOrPublic is true, the recipient will receive
	// the candidate with secret. The consequence of this is that
	// other participants will not be able to trace forward to the new
	// asset without any secret id. The recipient on the other hand
	// can freely choose whether other participants can trace back to the current
	// node or not.
	TransferAsset(
		ctx context.Context,
		requestTime time.Time,
		asset *model_sig_graph.Asset,
		ownerKey *model_sig_graph.UserKeyPair,
		peer *model_asset_transfer.Peer,
		exposedPrivateConnections map[string]model_asset_transfer.PrivateId,
		isNewConnectionSecretOrPublic bool,
	) (*model_asset_transfer.RequestToAcceptAsset, error)

	// - if isNewConnectionSecretOrPublic is true, the new node will
	// will reference back to the current node with a private edge.
	// The consequence of this is that other participants will not be
	// able to trace backward to the old asset without any secret id.
	AcceptRequestToAcceptAsset(
		ctx context.Context,
		peer *model_asset_transfer.Peer,
		request *model_asset_transfer.RequestToAcceptAsset,
		acceptOrReject bool,
		message string,
		isNewConnectionSecretOrPublic bool,
	) (updatedRequest *model_asset_transfer.RequestToAcceptAsset, newSecret string, oldSecret string, err error)

	SetNumberOfCandidatesSignature(ctx context.Context, numberOfCandidate uint32) error
}
