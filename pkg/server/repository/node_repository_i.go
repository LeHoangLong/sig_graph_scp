package repository_server

import (
	"context"
	"sig_graph_scp/pkg/model"
)

type NodeRepositoryI interface {
	UpsertNode(ctx context.Context, transactionId TransactionId, node *model.Node) error
	FetchNodesByOwnerPublicKey(ctx context.Context, transactionId TransactionId, ownerPublicKey string) ([]model.Node, error)
	FetchNodesByNodeId(ctx context.Context, transactionId TransactionId, namespace string, id map[model.NodeId]bool) ([]model.Node, error)
}
