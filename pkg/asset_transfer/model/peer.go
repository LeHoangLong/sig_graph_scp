package model_asset_transfer

import "sig_graph_scp/pkg/model"

type PeerProtocol struct {
	Type         model.EPeerProtocol `json:"protocol_type"`
	VersionMajor uint32              `json:"version_major"`
	VersionMinor uint32              `json:"version_minor"`
}

type Peer struct {
	Protocol      PeerProtocol `json:"protocol"`
	ConnectionUri string       `json:"connection_url"`

	PeerPemPublicKey string `json:"peer_pem_public_key"`
}
