package repository_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
)

type NodeRepositoryI interface {
	UpsertNode(ctx context.Context, transactionId TransactionId, node *model_server.Node) error
	FetchNodesByOwnerPublicKey(ctx context.Context, transactionId TransactionId, nodeType string, ownerPublicKey string) ([]model_server.Node, error)
	FetchNodesByNodeId(ctx context.Context, transactionId TransactionId, nodeType string, namespace string, id map[model_server.NodeId]bool) ([]model_server.Node, error)
}
