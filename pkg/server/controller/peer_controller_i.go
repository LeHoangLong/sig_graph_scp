package controller_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
	repository_server "sig_graph_scp/pkg/server/repository"
)

type PeerControllerI interface {
	GetPeersByUser(
		ctx context.Context,
		user *model_server.User,
		pagination repository_server.PaginationOption[model_server.PeerDbId],
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
