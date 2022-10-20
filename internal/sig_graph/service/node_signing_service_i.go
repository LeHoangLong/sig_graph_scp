package service_sig_graph

import (
	"context"
	"sig_graph_scp/pkg/model"
)

// return base64 encoded signature
type NodeSigningServiceI interface {
	Sign(ctx context.Context, userKeyPair *model.UserKeyPair, node any) (string, error)
}
