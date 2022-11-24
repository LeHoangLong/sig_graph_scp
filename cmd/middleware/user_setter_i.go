package middleware

import (
	model_server "sig_graph_scp/pkg/server/model"

	"github.com/gin-gonic/gin"
)

type UserSetterI interface {
	SetUser(c *gin.Context, user *model_server.User) error
}
