package repository_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
)

type GenericNodeRepositoryI interface {
	// returns ErrNotFound if no such generic node
	FetchNode(
		ctx context.Context,
		txId TransactionId,
		node *model_server.Node,
	) (fetchedNode model_server.Node, err error)

	UpsertNode(
		ctx context.Context,
		txId TransactionId,
		node *model_server.Node,
	) (savedNode model_server.Node, err error)
}
