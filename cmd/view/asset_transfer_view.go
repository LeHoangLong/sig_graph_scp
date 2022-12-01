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

type CreateRequestToAcceptAssetRequest struct {
	AssetId                        uint64     `json:"asset_id"`
	PeerId                         uint64     `json:"peer_id"`
	Edges                          []NodeEdge `json:"edges"`
	IsNewConnectionPublicOrPrivate bool       `json:"is_new_connection_public_or_private"`
}

func (v *assetTransferView) CreateRequestToAcceptAsset(c *gin.Context) {
	ctx := c.Request.Context()
	user := middleware.GetUser(c.Request.Context())

	request := CreateRequestToAcceptAssetRequest{}
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

	c.JSON(http.StatusCreated, requestToAcceptAsset)
	return
}

type GetReceivedRequestToAcceptAssetRequest struct {
	Status            string                 `form:"status"`
	OutboundOrInbound bool                   `form:"outbound_or_inbound"`
	MinId             model_server.RequestId `form:"min_id"`
	Limit             int                    `form:"limit"`
}

func (v *assetTransferView) GetReceivedRequestToAcceptAsset(c *gin.Context) {
	ctx := c.Request.Context()
	user := middleware.GetUser(c.Request.Context())

	request := GetReceivedRequestToAcceptAssetRequest{}
	if err := c.ShouldBind(&request); err != nil {
		utility.AbortBadRequest(c, err)
		return
	}

	pagination := repository_server.PaginationOption[model_server.RequestId]{
		Limit: request.Limit,
		MinId: request.MinId,
	}

	requestToAcceptAssets, err := v.controller.GetRequestsToAcceptAsset(
		ctx,
		user,
		request.Status,
		request.OutboundOrInbound,
		pagination,
	)

	if err != nil {
		utility.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, requestToAcceptAssets)
	return
}

type AcceptReceivedRequestToAcceptAssetRequest struct {
	RequestId             model_server.RequestId     `json:"request_id"`
	KeyPairId             model_server.UserKeyPairId `json:"key_id"`
	Accept                bool                       `json:"accept"`
	Message               string                     `json:"message"`
	IsNewConnectionSecret bool                       `json:"is_connection_secret"`
	ToInformSenderOfNewId bool                       `json:"to_inform_sender_of_new_id"`
}

func (v *assetTransferView) AcceptReceivedRequestToAcceptAsset(c *gin.Context) {
	ctx := c.Request.Context()
	user := middleware.GetUser(c.Request.Context())

	request := AcceptReceivedRequestToAcceptAssetRequest{}
	if err := c.ShouldBind(&request); err != nil {
		utility.AbortBadRequest(c, err)
		return
	}

	requestToAcceptAsset, err := v.controller.AcceptReceivedRequestsToAcceptAsset(
		ctx,
		user,
		request.KeyPairId,
		request.RequestId,
		request.Accept,
		request.Message,
		request.IsNewConnectionSecret,
		request.ToInformSenderOfNewId,
	)
	if err != nil {
		utility.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, requestToAcceptAsset)
	return
}

type GetPrivateEdgesRequest struct {
	RequestId uint64 `form:"request_id"`
}

func (v *assetTransferView) GetPrivateEdges(c *gin.Context) {
	ctx := c.Request.Context()
	user := middleware.GetUser(c.Request.Context())

	request := GetPrivateEdgesRequest{}
	if err := c.ShouldBind(&request); err != nil {
		utility.AbortBadRequest(c, err)
		return
	}

	relatedNodes, err := v.controller.FetchPrivateEdges(
		ctx,
		user,
		model_server.RequestId(request.RequestId),
	)
	if err != nil {
		utility.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, relatedNodes)
	return
}
