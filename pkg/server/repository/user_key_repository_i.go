package repository_server

import (
	"context"
	"sig_graph_scp/pkg/model"
)

type UserKeyRepositoryI interface {
	FetchKeyPairsOfUser(ctx context.Context, transactionId TransactionId, user *model.User) ([]model.UserKeyPair, error)
}
