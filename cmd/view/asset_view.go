package view

import (
	"net/http"
	"sig_graph_scp/cmd/middleware"
	utility "sig_graph_scp/cmd/utility"
	controller_server "sig_graph_scp/pkg/server/controller"
	model_server "sig_graph_scp/pkg/server/model"
	repository_server "sig_graph_scp/pkg/server/repository"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type assetView struct {
	controller controller_server.AssetControllerI
}

func NewAssetView(controller controller_server.AssetControllerI) *assetView {
	return &assetView{
		controller: controller,
	}
}

type CreateAssetRequest struct {
	MaterialName string          `json:"material_name"`
	Unit         string          `json:"unit"`
	Quantity     decimal.Decimal `json:"quantity"`
	KeyId        uint64          `json:"key_id"`
}

type SerializableNode struct {
}

func (v *assetView) CreateAsset(c *gin.Context) {
	user := middleware.GetUser(c.Request.Context())

	request := CreateAssetRequest{}
	if err := c.ShouldBind(&request); err != nil {
		utility.AbortBadRequest(c, err)
		return
	}

	asset, err := v.controller.CreateAsset(
		c.Request.Context(),
		user,
		request.MaterialName,
		request.Unit,
		request.Quantity,
		request.KeyId,
		[]model_server.Asset{},
		[]string{},
		[]string{},
		[]string{},
	)

	if err != nil {
		utility.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, asset)
	return
}

type GetAssetByIdRequest struct {
	AssetId  string `form:"asset_id"`
	UseCache bool   `form:"use_cache"`
}

func (v *assetView) GetAssetById(c *gin.Context) {
	user := middleware.GetUser(c.Request.Context())

	request := GetAssetByIdRequest{}
	if err := c.ShouldBind(&request); err != nil {
		utility.AbortBadRequest(c, err)
		return
	}

	asset, err := v.controller.GetAssetById(c.Request.Context(), user, model_server.NodeId(request.AssetId), request.UseCache)
	if err != nil {
		utility.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, asset)
	return
}

type GetAssetByDbIdRequest struct {
	Ids []uint64 `form:"ids"`
}

func (v *assetView) GetAssetByDbId(c *gin.Context) {
	user := middleware.GetUser(c.Request.Context())

	request := GetAssetByDbIdRequest{}
	if err := c.ShouldBind(&request); err != nil {
		utility.AbortBadRequest(c, err)
		return
	}

	idsMap := map[model_server.NodeDbId]bool{}
	for _, id := range request.Ids {
		idsMap[model_server.NodeDbId(id)] = true
	}

	assets, err := v.controller.GetAssetsFromCacheByDbId(
		c.Request.Context(),
		user,
		idsMap,
	)
	if err != nil {
		utility.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, assets)
	return
}

type GetOwnedAssetsFromCacheRequest struct {
	IsTransferred []bool                `form:"is_transferred"`
	MinId         model_server.NodeDbId `form:"min_id"`
	Limit         int                   `form:"limit"`
}

func (v *assetView) GetOwnedAssetsFromCache(
	c *gin.Context,
) {
	user := middleware.GetUser(c.Request.Context())

	request := GetOwnedAssetsFromCacheRequest{}
	if err := c.ShouldBind(&request); err != nil {
		utility.AbortBadRequest(c, err)
		return
	}

	pagination := repository_server.PaginationOption[model_server.NodeDbId]{
		MinId: request.MinId,
		Limit: request.Limit,
	}

	assets, err := v.controller.GetOwnedAssetsFromCache(
		c.Request.Context(),
		user,
		request.IsTransferred,
		pagination,
	)
	if err != nil {
		utility.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, assets)
	return
}
