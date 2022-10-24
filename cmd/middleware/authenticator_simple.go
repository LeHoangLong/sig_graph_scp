package middleware

import (
	model_server "sig_graph_scp/pkg/server/model"

	"github.com/gin-gonic/gin"
)

type authenticatorSimple struct {
}

func NewAuthenticatorSimple() *authenticatorSimple {
	return &authenticatorSimple{}
}

func (a *authenticatorSimple) Authenticate(c *gin.Context) {
	dummyUser := model_server.User{
		ID: 1,
	}
	ctx := setUser(c.Request.Context(), dummyUser)
	c.Request = c.Request.WithContext(ctx)
	c.Next()
}
