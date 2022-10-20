package controller_server

import (
	"context"
	"sig_graph_scp/pkg/model"
)

type UserKeyPairControllerI interface {
	FetchKeyPairsByUser(ctx context.Context, user *model.User) ([]model.UserKeyPair, error)
	AddKeyPairToUser(ctx context.Context, user *model.User, keyPair *model.UserKeyPair) error
}
