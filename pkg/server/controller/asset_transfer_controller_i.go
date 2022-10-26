package controller_server

import (
	"context"
	"sig_graph_scp/pkg/model"
	model_server "sig_graph_scp/pkg/server/model"
	repository_server "sig_graph_scp/pkg/server/repository"
)

type AssetTransferControllerI interface {
	TransferAsset(
		ctx context.Context,
		user *model_server.User,
		assetId model_server.NodeDbId,
		peerId model_server.PeerDbId,
		exposedSecretIds []repository_server.EdgeNodeId,
		isNewConnectionPublicOrPrivate bool,
	) (*model_server.RequestToAcceptAsset, error)

	GetReceivedRequestsToAcceptAsset(
		ctx context.Context,
		user *model_server.User,
		status model.ERequestToAcceptAssetStatus,
		pagination repository_server.PaginationOption[model_server.RequestId],
	) ([]model_server.RequestToAcceptAsset, error)

	/*

		GetSentRequestsToAcceptAsset(
			ctx context.Context,
			user *model_server.User,
			status model.ERequestToAcceptAssetStatus,
		) ([]model_server.RequestToAcceptAsset, error)

		AcceptReceivedRequestsToAcceptAsset(
			ctx context.Context,
			user *model_server.User,
			requestId model_server.RequestId,
		) (model_server.RequestToAcceptAsset, error)

		RejectReceivedRequestsToAcceptAsset(
			ctx context.Context,
			user *model_server.User,
			requestId model_server.RequestId,
			rejectMessage string,
		) (model_server.RequestToAcceptAsset, error)
	*/
}
