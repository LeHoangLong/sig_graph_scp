package service_sig_graph

import "context"

type NodeServiceI interface {
	DoNodeIdsExists(ctx context.Context, ids map[string]bool) (map[string]bool, error)
}
