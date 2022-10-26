package repository_server

import (
	"context"
	"sig_graph_scp/pkg/model"
	model_server "sig_graph_scp/pkg/server/model"
)

type AssetTransferRepositoryI interface {
	CreateAssetAcceptRequest(
		ctx context.Context,
		txId TransactionId,
		request *model_server.RequestToAcceptAsset,
	) error

	FetchAssetAcceptRequestsByUserAndStatus(
		ctx context.Context,
		txId TransactionId,
		user *model_server.User,
		status model.ERequestToAcceptAssetStatus,
		pagination PaginationOption[model_server.RequestId],
	) ([]model_server.RequestToAcceptAsset, error)
}
