package controller_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
	repository_server "sig_graph_scp/pkg/server/repository"
)

type nodeController struct {
	nodeRepository     repository_server.NodeRepositoryI
	transactionManager repository_server.TransactionManagerI
}

func NewNodeController(
	nodeRepository repository_server.NodeRepositoryI,
	transactionManager repository_server.TransactionManagerI,
) *nodeController {
	return &nodeController{
		nodeRepository:     nodeRepository,
		transactionManager: transactionManager,
	}
}

func (c *nodeController) UpdateNodeSecretId(
	ctx context.Context,
	node *model_server.Node,
	secretIds map[string]model_server.PrivateId,
) error {
	txId, err := c.transactionManager.BypassTransaction(ctx)
	if err != nil {
		return err
	}
	defer c.transactionManager.StopBypassedTransaction(ctx, txId)

	for hash := range node.PrivateChildrenIds {
		if privateId, ok := secretIds[hash]; ok {
			node.PrivateChildrenIds[hash] = privateId
		}
	}

	err = c.nodeRepository.UpsertNode(ctx, txId, node)
	if err != nil {
		return err
	}
	return nil
}
