package repository_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
)

type EdgeNodeId struct {
	Parent model_server.NodeId
	Child  model_server.NodeId
}

type NodeRepositoryI interface {
	UpsertNode(ctx context.Context, transactionId TransactionId, node *model_server.Node) error
	FetchNodesByOwnerPublicKey(ctx context.Context, transactionId TransactionId, nodeType string, ownerPublicKey string, pagination PaginationOption[model_server.NodeDbId]) ([]model_server.Node, error)
	FetchNodesByNodeId(ctx context.Context, transactionId TransactionId, nodeType string, namespace string, id map[model_server.NodeId]bool) ([]model_server.Node, error)
	FetchNodesByDbId(ctx context.Context, transactionId TransactionId, nodeType string, namespace string, id map[model_server.NodeDbId]bool) ([]model_server.Node, error)
	FetchPrivateEdgesByNodeIds(ctx context.Context, transactionId TransactionId, namespace string, edges []EdgeNodeId) ([]model_server.PrivateId, error)
}
