package view

import (
	"net/http"
	"sig_graph_scp/cmd/middleware"
	utility "sig_graph_scp/cmd/utility"
	controller_server "sig_graph_scp/pkg/server/controller"
	model_server "sig_graph_scp/pkg/server/model"
	repository_server "sig_graph_scp/pkg/server/repository"

	"github.com/gin-gonic/gin"
)

type assetTransferView struct {
	controller controller_server.AssetTransferControllerI
}

func NewAssetTransferView(controller controller_server.AssetTransferControllerI) *assetTransferView {
	return &assetTransferView{
		controller: controller,
	}
}

type NodeEdge struct {
	Parent string `json:"parent"`
	Child  string `json:"child"`
}

type CreateRequestToAcceptAsset struct {
	AssetId                        uint64     `json:"asset_id"`
	PeerId                         uint64     `json:"peer_id"`
	Edges                          []NodeEdge `json:"edges"`
	IsNewConnectionPublicOrPrivate bool       `json:"is_new_connection_public_or_private"`
}

func (v *assetTransferView) CreateRequestToAcceptAsset(c *gin.Context) {
	ctx := c.Request.Context()
	user := middleware.GetUser(c.Request.Context())

	request := CreateRequestToAcceptAsset{}
	if err := c.ShouldBindJSON(&request); err != nil {
		utility.AbortBadRequest(c, err)
		return
	}

	exposedSecretIds := make([]repository_server.EdgeNodeId, 0, len(request.Edges))
	for i := range request.Edges {
		exposedSecretIds = append(exposedSecretIds, repository_server.EdgeNodeId{
			Parent: model_server.NodeId(request.Edges[i].Parent),
			Child:  model_server.NodeId(request.Edges[i].Child),
		})
	}

	requestToAcceptAsset, err := v.controller.TransferAsset(
		ctx,
		user,
		model_server.NodeDbId(request.AssetId),
		request.PeerId,
		exposedSecretIds,
		request.IsNewConnectionPublicOrPrivate,
	)

	if err != nil {
		utility.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, requestToAcceptAsset)
	return
}
