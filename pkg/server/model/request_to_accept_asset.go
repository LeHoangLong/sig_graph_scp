package model_server

import "sig_graph_scp/pkg/model"

type RequestToAcceptAsset struct {
	Status                    model.ERequestToAcceptAssetStatus `json:"status"`
	IsOutboundOrInbound       bool                              `json:"is_outbound_or_inbound"`
	Time                      uint64                            `json:"time"`
	AckId                     string                            `json:"ack_id"`
	Accepted                  bool                              `json:"accepted"`
	AssetId                   DbId                              `json:"asset_id"`
	PeerId                    PeerDbId                          `json:"peer_id"`
	UserId                    UserId                            `json:"user_id"`
	ExposedPrivateConnections map[string]PrivateId              `json:"exposed_private_connections"`
}
