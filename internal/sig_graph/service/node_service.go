package service_sig_graph

import (
	"context"
	"encoding/json"
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

	data, err := s.smartContractService.Query("DoNodeIdsExists", string(idsJson))
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
