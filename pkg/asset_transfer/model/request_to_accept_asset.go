package model_asset_transfer

import (
	"sig_graph_scp/pkg/model"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"
)

type RequestToAcceptAsset struct {
	Status                    model.ERequestToAcceptAssetStatus `json:"status"`
	IsOutboundOrInbound       bool                              `json:"is_outbound_or_inbound"`
	TimeMs                    uint64                            `json:"time_ms"`
	AckId                     string                            `json:"ack_id"`
	Accepted                  bool                              `json:"accepted"`
	Asset                     model_sig_graph.Asset             `json:"asset"`
	NewAsset                  *model_sig_graph.Asset            `json:"new_asset"`
	PeerPemPublicKey          string                            `json:"peer_pem_public_key"`
	UserKeyPair               model_sig_graph.UserKeyPair       `json:"user_id"`
	ExposedPrivateConnections map[string]PrivateId              `json:"exposed_private_connections"`
	Candidates                []CandidateId                     `json:"candidates"`
}
