package service_server

import (
	"context"
	"fmt"
	"sig_graph_scp/pkg/model"
	model_server "sig_graph_scp/pkg/server/model"
	repository_server "sig_graph_scp/pkg/server/repository"
	api_sig_graph "sig_graph_scp/pkg/sig_graph/api"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"
	"sig_graph_scp/pkg/utility"
)

type NodeServiceI interface {
	UpdateNodeSecretId(
		ctx context.Context,
		txId repository_server.TransactionId,
		node *model_server.Node,
		secretIds map[string]model_server.PrivateId,
	) (*model_server.Node, error)

	// return model is in the model_server package
	FetchPrivateEdges(
		ctx context.Context,
		txId repository_server.TransactionId,
		exposedPrivateConnections map[string]model_server.PrivateId,
		endNode *model_server.Node,
		useCache bool,
	) (relatedNodes []any, err error)

	// retuns NotFound if any of id is not found
	// return model is in the model_server package
	FetchNodesByIds(
		ctx context.Context,
		txId repository_server.TransactionId,
		user *model_server.User,
		ids map[model_server.NodeId]bool,
		useCache bool,
	) (map[model_server.NodeId]any, error)
}

func NewNodeService(
	nodeRepository repository_server.NodeRepositoryI,
	genericNodeRepository map[model.ENodeType]repository_server.GenericNodeRepositoryI,
	sigGraphApi api_sig_graph.SigGraphClientApi,
) *nodeService {
	return &nodeService{
		nodeRepository: nodeRepository,
		sigGraphApi:    sigGraphApi,
	}
}

type nodeService struct {
	nodeRepository        repository_server.NodeRepositoryI
	genericNodeRepository map[model.ENodeType]repository_server.GenericNodeRepositoryI
	sigGraphApi           api_sig_graph.SigGraphClientApi
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

	for hash := range updatedNode.PrivateParentsIds {
		if privateId, ok := secretIds[hash]; ok {
			updatedNode.PrivateParentsIds[hash] = privateId
		}
	}

	err := s.nodeRepository.UpsertNode(ctx, txId, &updatedNode)
	if err != nil {
		return nil, err
	}
	return &updatedNode, nil
}

func (s *nodeService) FetchPrivateEdges(
	ctx context.Context,
	txId repository_server.TransactionId,
	exposedPrivateConnections map[string]model_server.PrivateId,
	endNode *model_server.Node,
	useCache bool,
) ([]any, error) {
	relatedNodesMap := map[model_server.NodeId]any{}
	err := s.fetchPrivateEdges(
		ctx,
		txId,
		endNode.Namespace,
		exposedPrivateConnections,
		endNode,
		relatedNodesMap,
		useCache,
	)
	if err != nil {
		return nil, err
	}

	ret := []any{}
	for id := range relatedNodesMap {
		ret = append(ret, relatedNodesMap[id])
	}

	return ret, nil
}

func (s *nodeService) FetchNodesByIds(
	ctx context.Context,
	txId repository_server.TransactionId,
	user *model_server.User,
	ids map[model_server.NodeId]bool,
	useCache bool,
) (map[model_server.NodeId]any, error) {
	namepsace := fmt.Sprintf("%d", user.ID)
	return s.fetchNodesByIds(
		ctx,
		txId,
		namepsace,
		ids,
		useCache,
	)
}

func (s *nodeService) fetchNodesByIds(
	ctx context.Context,
	txId repository_server.TransactionId,
	namespace string,
	ids map[model_server.NodeId]bool,
	useCache bool,
) (map[model_server.NodeId]any, error) {
	var err error
	cachedNodes := map[model_server.NodeId]any{}
	nodesToFetch := map[string]bool{}
	if len(ids) == 0 {
		return map[model_server.NodeId]any{}, nil
	}

	if useCache {
		cachedNodes, err = s.fetchCachedNodes(
			ctx,
			txId,
			namespace,
			ids,
		)
		if err != nil {
			return nil, err
		}
	}

	for id := range ids {
		if _, ok := cachedNodes[id]; !ok {
			nodesToFetch[string(id)] = true
		}
	}

	sigGraphNodes, err := s.sigGraphApi.FetchNodesByIds(
		ctx,
		nodesToFetch,
	)
	if err != nil {
		return nil, err
	}

	ret, err := s.saveGenericSigGraphNodeAndConvertToModel(
		ctx,
		txId,
		namespace,
		sigGraphNodes,
	)
	if err != nil {
		return nil, err
	}

	for id := range cachedNodes {
		ret[id] = cachedNodes[id]
	}

	return ret, nil
}

func (s *nodeService) fetchCachedNodes(
	ctx context.Context,
	txId repository_server.TransactionId,
	namespace string,
	ids map[model_server.NodeId]bool,
) (map[model_server.NodeId]any, error) {
	cachedNodes := map[model_server.NodeId]any{}
	nodes, err := s.nodeRepository.FetchNodesByNodeId(
		ctx,
		txId,
		repository_server.ENodeTypeAny,
		namespace,
		ids,
	)
	if err != nil {
		return nil, err
	}
	for i := range nodes {
		nodeType := nodes[i].NodeType
		cachedNodes[nodes[i].Id], err = s.genericNodeRepository[nodeType].FetchNode(
			ctx,
			txId,
			&nodes[i],
		)

		if err != nil {
			return nil, err
		}
	}

	return cachedNodes, nil
}

func (s *nodeService) saveGenericSigGraphNodeAndConvertToModel(
	ctx context.Context,
	txId repository_server.TransactionId,
	namespace string,
	sigGraphNodes map[string]any,
) (map[model_server.NodeId]any, error) {
	savedNodes := map[model_server.NodeId]any{}
	for id := range sigGraphNodes {
		var sigGraphNode model_sig_graph.Node
		var ok bool
		if sigGraphNode, ok = sigGraphNodes[id].(model_sig_graph.Node); !ok {
			return nil, utility.ErrInvalidState
		}

		nodeType := sigGraphNode.NodeType
		parsedNode, err := s.parseSigGraphNode(sigGraphNode, namespace)
		if err != nil {
			return nil, err
		}

		parsedNode, err = s.genericNodeRepository[nodeType].UpsertNode(
			ctx,
			txId,
			&parsedNode,
		)
		savedNodes[model_server.NodeId(id)] = parsedNode
	}

	return savedNodes, nil
}

func (s *nodeService) fetchPrivateEdges(
	ctx context.Context,
	txId repository_server.TransactionId,
	namespace string,
	exposedPrivateConnections map[string]model_server.PrivateId,
	node *model_server.Node,
	fetchedNodes map[model_server.NodeId]any,
	useCache bool,
) error {
	nodesToFetch := s.buildNodesToFetchMap(
		ctx,
		exposedPrivateConnections,
		node,
		fetchedNodes,
	)

	newlyFetchedNodes, err := s.fetchNodesByIds(
		ctx,
		txId,
		namespace,
		nodesToFetch,
		useCache,
	)
	if err != nil {
		return err
	}

	for id := range newlyFetchedNodes {
		fetchedNodes[model_server.NodeId(id)] = newlyFetchedNodes[id]
	}

	for id := range newlyFetchedNodes {
		if node, ok := newlyFetchedNodes[id].(model_server.Node); !ok {
			return utility.ErrInvalidState
		} else {
			err = s.fetchPrivateEdges(
				ctx,
				txId,
				namespace,
				exposedPrivateConnections,
				&node,
				fetchedNodes,
				useCache,
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// create model_server structs from model_sig_graph structs
// fields that cannot be filled will be filled with default values
func (s *nodeService) parseSigGraphNode(iNode any, namspace string) (any, error) {
	if node, ok := iNode.(model_sig_graph.Node); !ok {
		return model_server.Node{}, utility.ErrInvalidArgument
	} else {
		privateParentIds := map[string]model_server.PrivateId{}
		for hash := range node.PrivateParentsHashedIds {
			privateParentIds[hash] = model_server.PrivateId{}
		}

		privateChildrenIds := map[string]model_server.PrivateId{}
		for hash := range node.PrivateChildrenHashedIds {
			privateChildrenIds[hash] = model_server.PrivateId{}
		}
		modelNode := model_server.FromSigGraphNode(
			&node,
			0,
			namspace,
			privateParentIds,
			privateChildrenIds,
		)

		switch node.NodeType {
		case model.ENodeTypeAsset:
			if asset, ok := iNode.(model_sig_graph.Asset); !ok {
				return model_server.Node{}, utility.ErrInvalidArgument
			} else {
				modelAsset := model_server.FromSigGraphAsset(
					&asset,
					&modelNode,
				)
				return modelAsset, nil
			}
		default:
			return model_server.Node{}, utility.ErrInvalidArgument
		}
	}
}

func (s *nodeService) buildNodesToFetchMap(
	ctx context.Context,
	exposedPrivateConnections map[string]model_server.PrivateId,
	node *model_server.Node,
	fetchedNodes map[model_server.NodeId]any,
) map[model_server.NodeId]bool {
	nodesToFetch := map[model_server.NodeId]bool{}
	for hash := range node.PrivateParentsIds {
		if privateId, ok := exposedPrivateConnections[hash]; ok {
			if _, ok := fetchedNodes[privateId.ThisId]; !ok {
				nodesToFetch[privateId.ThisId] = true
			}
		}
	}

	for hash := range node.PrivateChildrenIds {
		if privateId, ok := exposedPrivateConnections[hash]; ok {
			if _, ok := fetchedNodes[privateId.ThisId]; !ok {
				nodesToFetch[privateId.ThisId] = true
			}
		}
	}

	return nodesToFetch
}
