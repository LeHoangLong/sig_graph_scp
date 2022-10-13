package middleware

import (
	"context"
	"sig_graph_scp/pkg/model"
)

type userCtxKeyType = string

const userCtxKey = "user"

func setUser(ctx context.Context, user model.User) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}

func GetUser(ctx context.Context) *model.User {
	user := ctx.Value(userCtxKey)
	if user == nil {
		return nil
	}

	if user, ok := user.(model.User); ok {
		return &user
	} else {
		return nil
	}
}
