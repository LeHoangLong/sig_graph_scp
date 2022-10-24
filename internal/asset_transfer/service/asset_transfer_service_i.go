package service_asset_transfer

import (
	"context"
	model_asset_transfer "sig_graph_scp/pkg/asset_transfer/model"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"
	"time"
)

type AssetTransferServiceI interface {
	TransferAsset(
		ctx context.Context,
		requestTime time.Time,
		asset *model_sig_graph.Asset,
		ownerKey *model_sig_graph.UserKeyPair,
		peer *model_asset_transfer.Peer,
		exposedPrivateConnections map[string]model_asset_transfer.PrivateId,
		isNewConnectionSecretOrPublic bool,
	) (*model_asset_transfer.RequestToAcceptAsset, error)

	SetNumberOfCandidatesSignature(ctx context.Context, numberOfCandidate uint32) error
}
