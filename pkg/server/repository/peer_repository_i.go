package repository_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
)

type PeerRepositoryI interface {
	FetchPeerById(ctx context.Context, txId TransactionId, id model_server.PeerDbId) (*model_server.Peer, error)
	FetchPeersByUser(ctx context.Context, txId TransactionId, user *model_server.User) ([]model_server.Peer, error)
	AddPeerToUser(ctx context.Context, txId TransactionId, peer *model_server.Peer) error
}
