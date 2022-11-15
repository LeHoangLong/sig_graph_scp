package controller_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
	repository_server "sig_graph_scp/pkg/server/repository"
	service_server "sig_graph_scp/pkg/server/service"
)

type nodeController struct {
	nodeService        service_server.NodeServiceI
	transactionManager repository_server.TransactionManagerI
}

func NewNodeController(
	nodeService service_server.NodeServiceI,
	transactionManager repository_server.TransactionManagerI,
) *nodeController {
	return &nodeController{
		nodeService:        nodeService,
		transactionManager: transactionManager,
	}
}

func (c *nodeController) UpdateNodeSecretId(
	ctx context.Context,
	node *model_server.Node,
	secretIds map[string]model_server.PrivateId,
) (*model_server.Node, error) {
	txId, err := c.transactionManager.BypassTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer c.transactionManager.StopBypassedTransaction(ctx, txId)

	return c.nodeService.UpdateNodeSecretId(ctx, txId, node, secretIds)
}

func (c *nodeController) FetchPrivateEdges(
	ctx context.Context,
	exposedPrivateConnections map[string]model_server.PrivateId,
	endNode *model_server.Node,
	useCache bool,
) (relatedNodes []any, err error) {
	txId, err := c.transactionManager.BypassTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer c.transactionManager.StopBypassedTransaction(ctx, txId)

	return c.nodeService.FetchPrivateEdges(ctx, txId, exposedPrivateConnections, endNode, useCache)
}
