package service_sig_graph

import "sig_graph_scp/pkg/model"

type IdGenerateServiceI interface {
	NewFullId() model.NodeId
}
