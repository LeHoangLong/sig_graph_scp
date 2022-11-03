package service_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
	repository_server "sig_graph_scp/pkg/server/repository"
)

type NodeServiceI interface {
	UpdateNodeSecretId(
		ctx context.Context,
		txId repository_server.TransactionId,
		node *model_server.Node,
		secretIds map[string]model_server.PrivateId,
	) (*model_server.Node, error)
}

func NewNodeService(
	nodeRepository repository_server.NodeRepositoryI,
) *nodeService {
	return &nodeService{
		nodeRepository: nodeRepository,
	}
}

type nodeService struct {
	nodeRepository repository_server.NodeRepositoryI
}

func (s *nodeService) UpdateNodeSecretId(
	ctx context.Context,
	txId repository_server.TransactionId,
	node *model_server.Node,
	secretIds map[string]model_server.PrivateId,
) (*model_server.Node, error) {
	var updatedNode model_server.Node = *node
	for hash := range updatedNode.PrivateChildrenIds {
		if privateId, ok := secretIds[hash]; ok {
			updatedNode.PrivateChildrenIds[hash] = privateId
		}
	}

	err := s.nodeRepository.UpsertNode(ctx, txId, &updatedNode)
	if err != nil {
		return nil, err
	}
	return &updatedNode, nil
}
