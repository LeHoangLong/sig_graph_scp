package view

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sig_graph_scp/cmd/middleware"
	"sig_graph_scp/cmd/utility"
	controller_server "sig_graph_scp/pkg/server/controller"
	model_server "sig_graph_scp/pkg/server/model"
	repository_server "sig_graph_scp/pkg/server/repository"

	"github.com/gin-gonic/gin"
)

type userKeyPairView struct {
	controller controller_server.UserKeyPairControllerI
}

func NewUserKeyPairView(
	controller controller_server.UserKeyPairControllerI,
) *userKeyPairView {
	return &userKeyPairView{
		controller: controller,
	}
}

type GetUserKeyPairsByUserRequest struct {
	MinId model_server.UserKeyPairId `form:"min_id"`
	Limit int                        `form:"limit"`
}

func (v *userKeyPairView) GetUserKeyPairsByUser(c *gin.Context) {
	user := middleware.GetUser(c.Request.Context())

	request := GetUserKeyPairsByUserRequest{}
	if err := c.ShouldBind(&request); err != nil {
		utility.AbortBadRequest(c, err)
		return
	}

	pagination := repository_server.PaginationOption[model_server.PeerDbId]{
		MinId: request.MinId,
		Limit: request.Limit,
	}
	userKeyPairs, err := v.controller.FetchKeyPairsByUser(c.Request.Context(), user, pagination)
	if err != nil {
		utility.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, userKeyPairs)
	return
}

type AddUserKeyPairToUserRequest struct {
	PublicKey  *multipart.FileHeader `form:"public_key"`
	PrivateKey *multipart.FileHeader `form:"private_key"`
}

func (v *userKeyPairView) AddUserKeyPairToUser(c *gin.Context) {
	user := middleware.GetUser(c.Request.Context())
	request := AddUserKeyPairToUserRequest{}
	if err := c.ShouldBind(&request); err != nil {
		utility.AbortBadRequest(c, err)
		return
	}

	if request.PrivateKey == nil || request.PublicKey == nil {
		utility.AbortBadRequest(c, fmt.Errorf("missing parameter"))
		return
	}

	publicKey := ""
	{
		file, err := request.PublicKey.Open()
		if err != nil {
			utility.AbortWithError(c, err)
			return
		}

		data, err := io.ReadAll(file)
		if err != nil {
			utility.AbortWithError(c, err)
			return
		}
		publicKey = string(data)
	}

	privateKey := ""
	{
		file, err := request.PrivateKey.Open()
		if err != nil {
			utility.AbortWithError(c, err)
			return
		}

		data, err := io.ReadAll(file)
		if err != nil {
			utility.AbortWithError(c, err)
			return
		}
		privateKey = string(data)
	}

	userKeyPair := &model_server.UserKeyPair{
		Private: privateKey,
		Public:  publicKey,
	}

	err := v.controller.AddKeyPairToUser(c.Request.Context(), user, userKeyPair)
	if err != nil {
		utility.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, "")
	return
}
