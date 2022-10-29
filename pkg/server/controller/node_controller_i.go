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
	) error
}
