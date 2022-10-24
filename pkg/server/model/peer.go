package model_server

import "sig_graph_scp/pkg/model"

type PeerDbId = uint64

type PeerProtocol struct {
	Type         model.EPeerProtocol `json:"protocol_type"`
	VersionMajor uint32              `json:"version_major"`
	VersionMinor uint32              `json:"version_minor"`
}

type Peer struct {
	UserId UserId `json:"user_id"`

	DbId PeerDbId `json:"db_id"`

	Protocol      PeerProtocol `json:"protocol"`
	ConnectionUri string       `json:"connection_url"`

	PeerPemPublicKey string `json:"peer_pem_public_key"`
}
