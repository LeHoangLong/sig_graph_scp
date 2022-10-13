package middleware

import (
	"sig_graph_scp/pkg/model"

	"github.com/gin-gonic/gin"
)

type authenticatorSimple struct {
}

func NewAuthenticatorSimple() *authenticatorSimple {
	return &authenticatorSimple{}
}

func (a *authenticatorSimple) Authenticate(c *gin.Context) {
	dummyUser := model.User{
		ID: 1,
	}
	ctx := setUser(c.Request.Context(), dummyUser)
	c.Request = c.Request.WithContext(ctx)
	c.Next()
}
