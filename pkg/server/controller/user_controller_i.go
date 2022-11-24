package controller_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
)

type UserControllerI interface {
	// return ErrAlreadyExists if username already exists
	CreateUserWithUsernameAndPassword(ctx context.Context, username string, password string) (model_server.User, error)
	// return ErrNotFound if username does not exist or password does not match
	FetchUserWithusernameAndPassword(ctx context.Context, username string, password string) (model_server.User, error)
}
