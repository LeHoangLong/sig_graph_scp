package middleware

import (
	"github.com/gin-gonic/gin"
)

type UserUnsetterI interface {
	UnsetUser(c *gin.Context) error
}
