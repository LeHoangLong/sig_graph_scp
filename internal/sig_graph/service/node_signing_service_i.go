package service_sig_graph

import (
	"context"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"
)

// return base64 encoded signature
type NodeSigningServiceI interface {
	Sign(ctx context.Context, userKeyPair *model_sig_graph.UserKeyPair, node any) (string, error)
}
