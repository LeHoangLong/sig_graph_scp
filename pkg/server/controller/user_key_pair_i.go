package controller_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
	repository_server "sig_graph_scp/pkg/server/repository"
)

type UserKeyPairControllerI interface {
	FetchKeyPairsByUser(ctx context.Context, user *model_server.User, pagination repository_server.PaginationOption[model_server.UserKeyPairId]) ([]model_server.UserKeyPair, error)
	AddKeyPairToUser(ctx context.Context, user *model_server.User, keyPair *model_server.UserKeyPair) error
}
