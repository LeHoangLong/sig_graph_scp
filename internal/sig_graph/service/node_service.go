package service_sig_graph

import (
	"context"
	"encoding/json"
	"sig_graph_scp/pkg/model"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"
	"sig_graph_scp/pkg/utility"
)

type nodeService struct {
	smartContractService SmartContractServiceI
}

func NewNodeService(smartContractService SmartContractServiceI) NodeServiceI {
	return &nodeService{
		smartContractService: smartContractService,
	}
}

func (s *nodeService) DoNodeIdsExists(
	ctx context.Context,
	ids map[string]bool,
) (map[string]bool, error) {
	idsJson, err := json.Marshal(ids)
	if err != nil {
		return nil, err
	}

	data, err := s.smartContractService.Query("DoNodeIdsExist", string(idsJson))
	if err != nil {
		return nil, err
	}

	exist := map[string]bool{}
	err = json.Unmarshal([]byte(data), &exist)
	if err != nil {
		return nil, err
	}

	return exist, nil
}

type getNodesByIdRequest struct {
	Ids map[string]bool `json:"ids"`
}

// return NotFound if any one id is not found
func (s *nodeService) FetchNodesByIds(ctx context.Context, ids map[string]bool) (map[string]any, error) {
	request := getNodesByIdRequest{}
	request.Ids = ids

	requestJson, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	responseJson, err := s.smartContractService.Query("GetNodesById", string(requestJson))
	if err != nil {
		return nil, err
	}

	response := map[string]map[string]any{}
	err = json.Unmarshal([]byte(responseJson), &response)
	if err != nil {
		return nil, err
	}

	parsedNodes, err := s.parseNodeType(
		ctx,
		response,
	)
	if err != nil {
		return nil, err
	}

	return parsedNodes, nil
}

func (s *nodeService) parseNodeType(
	ctx context.Context,
	nodes map[string]map[string]any,
) (map[string]any, error) {
	ret := map[string]any{}
	for id := range nodes {
		nodeJson, err := json.Marshal(nodes[id])
		if err != nil {
			return nil, err
		}

		nodeType := nodes[id]["type"]
		switch nodeType {
		case model.ENodeTypeAsset:
			asset := model_sig_graph.Asset{}
			err := json.Unmarshal(nodeJson, &asset)
			if err != nil {
				return nil, err
			}

			ret[id] = asset
		default:
			return nil, utility.ErrInvalidState
		}
	}

	return ret, nil
}
