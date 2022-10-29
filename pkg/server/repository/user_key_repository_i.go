package repository_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
)

type UserKeyRepositoryI interface {
	FetchKeyPairsOfUser(ctx context.Context, transactionId TransactionId, user *model_server.User, pagination PaginationOption[model_server.UserKeyPairId]) ([]model_server.UserKeyPair, error)
	AddKeyPairToUser(ctx context.Context, transactionId TransactionId, user *model_server.User, keyPair *model_server.UserKeyPair) error
	FetchUserWithPublicKey(ctx context.Context, transactionId TransactionId, publicKey string) (*model_server.User, error)
	FetchKeyPairsByIds(ctx context.Context, transactionId TransactionId, user *model_server.User, ids map[model_server.UserKeyPairId]bool) ([]model_server.UserKeyPair, error)
}
