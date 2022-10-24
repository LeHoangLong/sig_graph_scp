package controller_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
)

type UserKeyPairControllerI interface {
	FetchKeyPairsByUser(ctx context.Context, user *model_server.User) ([]model_server.UserKeyPair, error)
	AddKeyPairToUser(ctx context.Context, user *model_server.User, keyPair *model_server.UserKeyPair) error
}
