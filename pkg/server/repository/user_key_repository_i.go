package repository_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
)

type UserKeyRepositoryI interface {
	FetchKeyPairsOfUser(ctx context.Context, transactionId TransactionId, user *model_server.User) ([]model_server.UserKeyPair, error)
	AddKeyPairToUser(ctx context.Context, transactionId TransactionId, user *model_server.User, keyPair *model_server.UserKeyPair) error
}
