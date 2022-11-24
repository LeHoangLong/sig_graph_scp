package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	model_server "sig_graph_scp/pkg/server/model"

	"github.com/gin-gonic/gin"
)

type authenticatorSimple struct {
	maxAge_s int
	domain   string
}

func NewAuthenticatorSimple(
	maxAge_s int,
	domain string,
) *authenticatorSimple {
	return &authenticatorSimple{
		maxAge_s: maxAge_s,
		domain:   domain,
	}
}

const userCookieName = "user"

func (a *authenticatorSimple) Authenticate(c *gin.Context) {
	userCookie, err := c.Request.Cookie(userCookieName)
	if err != nil {
		c.AbortWithError(http.StatusForbidden, fmt.Errorf("login cookie not found"))
		return
	}

	user := model_server.User{}
	userStr, err := url.QueryUnescape(userCookie.Value)
	if err != nil {
		c.AbortWithError(http.StatusForbidden, fmt.Errorf("%s", "invalid cookie"))
		return
	}

	err = json.Unmarshal([]byte(userStr), &user)
	if err != nil {
		c.AbortWithError(http.StatusForbidden, fmt.Errorf("%s", "invalid cookie"))
		return
	}

	ctx := setUser(c.Request.Context(), user)
	c.Request = c.Request.WithContext(ctx)
	c.Next()
}

func (a *authenticatorSimple) SetUser(c *gin.Context, user *model_server.User) error {
	userStr, err := json.Marshal(user)
	if err != nil {
		return err
	}

	c.SetCookie(userCookieName, string(userStr), a.maxAge_s, "/", a.domain, true, true)
	return nil
}

func (a *authenticatorSimple) UnsetUser(c *gin.Context) error {
	c.SetCookie(userCookieName, "", -1, "", a.domain, true, true)
	return nil
}
