package service_sig_graph

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type idGenerateServiceUuid struct {
	graphName string
}

func NewIdGenerateServiceUuid(graphName string) *idGenerateServiceUuid {
	return &idGenerateServiceUuid{graphName: graphName}
}

func (s *idGenerateServiceUuid) NewFullId(ctx context.Context) (string, error) {
	return fmt.Sprintf("%s:%s", s.graphName, uuid.New().String()), nil
}
