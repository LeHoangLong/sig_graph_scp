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
		user *model_server.User,
		exposedPrivateConnections map[string]model_server.PrivateId,
		endNode *model_server.Node,
		useCache bool,
	) (relatedNodes []model_server.Node, err error)

	// retuns NotFound if any of id is not found
	// return model is in the model_server package
	FetchNodesByIds(
		ctx context.Context,
		txId repository_server.TransactionId,
		user *model_server.User,
		ids map[model_server.NodeId]bool,
		useCache bool,
	) (map[model_server.NodeId]model_server.Node, error)
}

func NewNodeService(
	nodeRepository repository_server.NodeRepositoryI,
	genericNodeRepository map[model.ENodeType]repository_server.GenericNodeRepositoryI,
	sigGraphApi api_sig_graph.SigGraphClientApi,
) *nodeService {
	return &nodeService{
		nodeRepository:        nodeRepository,
		sigGraphApi:           sigGraphApi,
		genericNodeRepository: genericNodeRepository,
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
	user *model_server.User,
	exposedPrivateConnections map[string]model_server.PrivateId,
	endNode *model_server.Node,
	useCache bool,
) ([]model_server.Node, error) {
	reversedIds := map[string]model_server.PrivateId{}
	relatedNodesMap := map[model_server.NodeId]model_server.Node{}
	// reverse the exposed connections so that we also fetch them
	for _, privateId := range exposedPrivateConnections {
		reversedId := model_server.ReversePrivateId(&privateId)
		reversedIds[reversedId.ThisHash] = reversedId
	}

	for hash := range reversedIds {
		exposedPrivateConnections[hash] = reversedIds[hash]
	}

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

	relatedNodesMapWithStringKey := map[string]model_server.Node{
		string(endNode.Id): *endNode,
	}
	for id := range relatedNodesMap {
		relatedNodesMapWithStringKey[string(id)] = relatedNodesMap[id]
	}

	namespace := fmt.Sprintf("%d", user.ID)
	savedNodes, err := s.saveGenericModel(
		ctx,
		txId,
		namespace,
		relatedNodesMapWithStringKey,
	)
	if err != nil {
		return nil, err
	}

	ret := []model_server.Node{}
	for id := range savedNodes {
		ret = append(ret, savedNodes[id])
	}
	return ret, nil
}

func (s *nodeService) FetchNodesByIds(
	ctx context.Context,
	txId repository_server.TransactionId,
	user *model_server.User,
	ids map[model_server.NodeId]bool,
	useCache bool,
) (map[model_server.NodeId]model_server.Node, error) {
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
) (map[model_server.NodeId]model_server.Node, error) {
	var err error
	cachedNodes := map[model_server.NodeId]model_server.Node{}
	nodesToFetch := map[string]bool{}
	if len(ids) == 0 {
		return map[model_server.NodeId]model_server.Node{}, nil
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

	savedNodes, err := s.saveGenericSigGraphNodeAndConvertToModel(
		ctx,
		txId,
		namespace,
		sigGraphNodes,
	)
	if err != nil {
		return nil, err
	}

	ret := cachedNodes
	for id := range savedNodes {
		ret[id] = savedNodes[id]
	}

	return ret, nil
}

func (s *nodeService) fetchCachedNodes(
	ctx context.Context,
	txId repository_server.TransactionId,
	namespace string,
	ids map[model_server.NodeId]bool,
) (map[model_server.NodeId]model_server.Node, error) {
	extendedNodes := map[model_server.NodeId]model_server.Node{}
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
		extendedNodes[nodes[i].Id], err = s.genericNodeRepository[nodeType].FetchNode(
			ctx,
			txId,
			&nodes[i],
		)

		if err != nil {
			return nil, err
		}
	}

	return extendedNodes, nil
}

func (s *nodeService) saveGenericModel(
	ctx context.Context,
	txId repository_server.TransactionId,
	namespace string,
	modelNodes map[string]model_server.Node,
) (map[model_server.NodeId]model_server.Node, error) {
	savedNodes := map[model_server.NodeId]model_server.Node{}
	for id := range modelNodes {
		_, nodeType, err := s.extractServerNode(modelNodes[id])
		if err != nil {
			return nil, err
		}
		modelNode := modelNodes[id]

		savedNode, err := s.genericNodeRepository[nodeType].UpsertNode(
			ctx,
			txId,
			&modelNode,
		)
		if err != nil {
			return nil, err
		}
		savedNodes[model_server.NodeId(id)] = savedNode
	}

	return savedNodes, nil
}

func (s *nodeService) saveGenericSigGraphNodeAndConvertToModel(
	ctx context.Context,
	txId repository_server.TransactionId,
	namespace string,
	sigGraphNodes map[string]any,
) (map[model_server.NodeId]model_server.Node, error) {
	savedNodes := map[model_server.NodeId]model_server.Node{}
	for id := range sigGraphNodes {
		parsedNode, nodeType, err := s.parseSigGraphNodeToServerNode(sigGraphNodes[id], namespace)
		if err != nil {
			return nil, err
		}

		parsedNode, err = s.genericNodeRepository[nodeType].UpsertNode(
			ctx,
			txId,
			&parsedNode,
		)

		if err != nil {
			return nil, err
		}
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
	fetchedNodes map[model_server.NodeId]model_server.Node,
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
		for hash := range exposedPrivateConnections {
			exposedPrivateId := exposedPrivateConnections[hash]
			otherHash := exposedPrivateConnections[hash].OtherHash
			if _, ok := node.PrivateParentsIds[otherHash]; ok {
				node.PrivateParentsIds[otherHash] = model_server.ReversePrivateId(&exposedPrivateId)
			}

			thisHash := exposedPrivateConnections[hash].ThisHash
			if _, ok := node.PrivateParentsIds[thisHash]; ok {
				node.PrivateParentsIds[thisHash] = exposedPrivateId
			}
		}

		for hash := range exposedPrivateConnections {
			exposedPrivateId := exposedPrivateConnections[hash]
			otherHash := exposedPrivateConnections[hash].OtherHash
			if _, ok := node.PrivateChildrenIds[otherHash]; ok {
				node.PrivateChildrenIds[otherHash] = model_server.ReversePrivateId(&exposedPrivateId)
			}

			thisHash := exposedPrivateConnections[hash].ThisHash
			if _, ok := node.PrivateChildrenIds[thisHash]; ok {
				node.PrivateChildrenIds[thisHash] = exposedPrivateId
			}
		}

		fetchedNode, _, err := s.extractServerNode(newlyFetchedNodes[id])
		if err != nil {
			return err
		}
		err = s.fetchPrivateEdges(
			ctx,
			txId,
			namespace,
			exposedPrivateConnections,
			&fetchedNode,
			fetchedNodes,
			useCache,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *nodeService) extractServerNode(iNode any) (extractedNode model_server.Node, nodeType model.ENodeType, err error) {
	if node, ok := iNode.(model_server.Node); ok {
		if asset, ok := node.Extra.(*model_server.Asset); ok {
			return asset.Node, model.ENodeTypeAsset, nil
		}
		return node, model.ENodeTypeNode, nil
	} else if node, ok := iNode.(model_server.Asset); ok {
		return node.Node, model.ENodeTypeAsset, nil
	} else {
		return model_server.Node{}, "", utility.ErrInvalidArgument
	}
}

func (s *nodeService) extractSigGraphNode(iNode any) (extractedNode model_sig_graph.Node, err error) {
	if node, ok := iNode.(model_sig_graph.Node); ok {
		return node, nil
	} else if node, ok := iNode.(model_sig_graph.Asset); ok {
		return node.Node, nil
	} else {
		return model_sig_graph.Node{}, utility.ErrInvalidArgument
	}
}

// create model_server structs from model_sig_graph structs
// fields that cannot be filled will be filled with default values
func (s *nodeService) parseSigGraphNodeToServerNode(iNode any, namspace string) (modelServerNode model_server.Node, nodeType model.ENodeType, err error) {
	extractedNode, err := s.extractSigGraphNode(iNode)
	if err != nil {
		return model_server.Node{}, "", err
	}

	privateParentIds := map[string]model_server.PrivateId{}
	for hash := range extractedNode.PrivateParentsHashedIds {
		privateParentIds[hash] = model_server.PrivateId{}
	}

	privateChildrenIds := map[string]model_server.PrivateId{}
	for hash := range extractedNode.PrivateChildrenHashedIds {
		privateChildrenIds[hash] = model_server.PrivateId{}
	}
	modelNode := model_server.FromSigGraphNode(
		&extractedNode,
		0,
		namspace,
		privateParentIds,
		privateChildrenIds,
	)

	switch extractedNode.NodeType {
	case model.ENodeTypeAsset:
		if asset, ok := iNode.(model_sig_graph.Asset); !ok {
			return model_server.Node{}, "", utility.ErrInvalidArgument
		} else {
			modelAsset := model_server.FromSigGraphAsset(
				&asset,
				&modelNode,
			)
			return modelAsset.Node, model.ENodeTypeAsset, nil
		}
	default:
		return model_server.Node{}, "", utility.ErrInvalidArgument
	}
}

func (s *nodeService) buildNodesToFetchMap(
	ctx context.Context,
	exposedPrivateConnections map[string]model_server.PrivateId,
	node *model_server.Node,
	fetchedNodes map[model_server.NodeId]model_server.Node,
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
