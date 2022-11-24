package view

import (
	"net/http"
	"sig_graph_scp/cmd/middleware"
	utility "sig_graph_scp/cmd/utility"
	controller_server "sig_graph_scp/pkg/server/controller"

	"github.com/gin-gonic/gin"
)

type userView struct {
	controller   controller_server.UserControllerI
	userSetter   middleware.UserSetterI
	userUnsetter middleware.UserUnsetterI
}

func NewUserView(
	controller controller_server.UserControllerI,
	userSetter middleware.UserSetterI,
	userUnsetter middleware.UserUnsetterI,
) *userView {
	return &userView{
		controller:   controller,
		userSetter:   userSetter,
		userUnsetter: userUnsetter,
	}
}

type LogInRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (v *userView) LogIn(c *gin.Context) {
	request := LogInRequest{}
	if err := c.ShouldBind(&request); err != nil {
		utility.AbortBadRequest(c, err)
		return
	}

	user, err := v.controller.FetchUserWithusernameAndPassword(
		c.Request.Context(),
		request.Username,
		request.Password,
	)

	if err != nil {
		utility.AbortWithError(c, err)
		return
	}

	err = v.userSetter.SetUser(c, &user)
	if err != nil {
		utility.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
	return
}

type SignUpRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (v *userView) SignUp(c *gin.Context) {
	request := SignUpRequest{}
	if err := c.ShouldBind(&request); err != nil {
		utility.AbortBadRequest(c, err)
		return
	}

	user, err := v.controller.CreateUserWithUsernameAndPassword(
		c.Request.Context(),
		request.Username,
		request.Password,
	)
	if err != nil {
		utility.AbortWithError(c, err)
		return
	}

	err = v.userSetter.SetUser(c, &user)
	if err != nil {
		utility.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, user)
	return
}

func (v *userView) LogOut(c *gin.Context) {
	err := v.userUnsetter.UnsetUser(c)

	if err != nil {
		utility.AbortWithError(c, err)
		return
	}

	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(""))
	return
}
