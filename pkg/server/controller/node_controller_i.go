package controller_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
)

type NodeControllerI interface {
	UpdateNodeSecretId(
		ctx context.Context,
		node *model_server.Node,
		secretIds map[string]model_server.PrivateId,
	) (*model_server.Node, error)

	FetchPrivateEdges(
		ctx context.Context,
		user *model_server.User,
		exposedPrivateConnections map[string]model_server.PrivateId,
		endNode *model_server.Node,
		useCache bool,
	) (relatedNodes []model_server.Node, err error)
}
