package view

import (
	"net/http"
	"sig_graph_scp/cmd/middleware"
	"sig_graph_scp/pkg/model"
	controller_server "sig_graph_scp/pkg/server/controller"

	"github.com/gin-gonic/gin"
)

type assetView struct {
	controller controller_server.AssetControllerI
}

func NewAssetView(controller controller_server.AssetControllerI) *assetView {
	return &assetView{
		controller: controller,
	}
}

type GetAssetByIdRequest struct {
	AssetId  string `json:"asset_id"`
	UseCache bool   `json:"use_cache"`
}

func (v *assetView) GetAssetById(c *gin.Context) {
	user := middleware.GetUser(c.Request.Context())
	if user == nil {
		c.JSON(http.StatusBadRequest, "missing user")
		return
	}

	request := GetAssetByIdRequest{}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	asset, err := v.controller.GetAssetById(c.Request.Context(), user, model.NodeId(request.AssetId), request.UseCache)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, asset)
	return
}
