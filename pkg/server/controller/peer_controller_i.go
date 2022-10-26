package controller_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
)

type PeerControllerI interface {
	GetPeersByUser(
		ctx context.Context,
		user *model_server.User,
	) ([]model_server.Peer, error)

	AddPeerToUser(
		ctx context.Context,
		user *model_server.User,
		protocolType string,
		versionMajor uint32,
		versionMinor uint32,
		connectionUri string,
		peerPemPublicKey string,
	) (*model_server.Peer, error)
}
