package middleware

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
)

type userCtxKeyType = string

const userCtxKey = "user"

func setUser(ctx context.Context, user model_server.User) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}

func GetUser(ctx context.Context) *model_server.User {
	user := ctx.Value(userCtxKey)
	if user == nil {
		return nil
	}

	if user, ok := user.(model_server.User); ok {
		return &user
	} else {
		return nil
	}
}
