package service_sig_graph

import (
	"context"
)

type IdGenerateServiceI interface {
	NewFullId(ctx context.Context) (string, error)
}
