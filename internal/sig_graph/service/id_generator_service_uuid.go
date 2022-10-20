package service_sig_graph

import (
	"fmt"
	"sig_graph_scp/pkg/model"

	"github.com/google/uuid"
)

type idGenerateServiceUuid struct {
	graphName string
}

func NewIdGenerateServiceUuid(graphName string) *idGenerateServiceUuid {
	return &idGenerateServiceUuid{graphName: graphName}
}

func (s *idGenerateServiceUuid) NewFullId() model.NodeId {
	return model.NodeId(fmt.Sprintf("%s:%s", s.graphName, uuid.New().String()))
}
