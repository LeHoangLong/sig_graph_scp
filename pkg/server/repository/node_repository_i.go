package repository_server

import (
	"context"
	"sig_graph_scp/pkg/model"
	model_server "sig_graph_scp/pkg/server/model"
)

type EdgeNodeId struct {
	Parent model_server.NodeId
	Child  model_server.NodeId
}

const ENodeTypeAny model.ENodeType = "any"

// TODO: Add method to remove edge
//
//	NOTE: An edge is made up of the source and target vertex. This is different\
//		from the parent-child relationship expressed in the SigGraph model. For example,
//		an edge may be starting from either the child or the parent and the source will correspond
//		to the starting point respectively.
//
// NOTE: An unknown edge is one where the target vertex is unknown (either the this_id or this_hash string is empty)
type NodeRepositoryI interface {
	// if existing node have more private id than iNode,
	// then those extra edges won't be affected. The edges
	// whose target vertex are known are also not affected.
	// As a result, this function can only add new known / unknown
	// edges or add vertex information to an existing unknown
	// edge. You must use other methods if you want to delete an existing
	// edge.
	UpsertNode(ctx context.Context, transactionId TransactionId, iNode *model_server.Node) error
	FetchNodesByOwnerPublicKey(ctx context.Context, transactionId TransactionId, nodeType model.ENodeType, namespace string, ownerPublicKey string, pagination PaginationOption[model_server.NodeDbId]) ([]model_server.Node, error)
	FetchNodesByNodeId(ctx context.Context, transactionId TransactionId, nodeType model.ENodeType, namespace string, id map[model_server.NodeId]bool) ([]model_server.Node, error)
	FetchNodesByDbId(ctx context.Context, transactionId TransactionId, nodeType model.ENodeType, namespace string, id map[model_server.NodeDbId]bool) ([]model_server.Node, error)
	FetchPrivateEdgesByNodeIds(ctx context.Context, transactionId TransactionId, namespace string, edges []EdgeNodeId) ([]model_server.PrivateId, error)
}
