package model_server

import (
	model_asset_transfer "sig_graph_scp/pkg/asset_transfer/model"
	"sig_graph_scp/pkg/model"
)

type PeerDbId = uint64

type PeerProtocol struct {
	Type         model.EPeerProtocol `json:"protocol_type"`
	VersionMajor uint32              `json:"version_major"`
	VersionMinor uint32              `json:"version_minor"`
}

type Peer struct {
	PeerDbId PeerDbId `json:"db_id"`

	UserId UserId `json:"user_id"`

	Protocol      PeerProtocol `json:"protocol"`
	ConnectionUri string       `json:"connection_uri"`

	PeerPemPublicKey string `json:"peer_pem_public_key"`
	Name             string `json:"name"`
}

func ToAssetTransferPeerProtocol(protocol *PeerProtocol) model_asset_transfer.PeerProtocol {
	return model_asset_transfer.PeerProtocol{
		Type:         protocol.Type,
		VersionMajor: protocol.VersionMajor,
		VersionMinor: protocol.VersionMinor,
	}
}

func ToAssetTransferPeer(peer *Peer) model_asset_transfer.Peer {
	return model_asset_transfer.Peer{
		Protocol:         ToAssetTransferPeerProtocol(&peer.Protocol),
		ConnectionUri:    peer.ConnectionUri,
		PeerPemPublicKey: peer.PeerPemPublicKey,
	}
}
