package view

import (
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

type peerView struct {
	controller controller_server.PeerControllerI
}

func NewPeerView(controller controller_server.PeerControllerI) *peerView {
	return &peerView{
		controller: controller,
	}
}

type GetPeerRequest struct {
	MinId model_server.PeerDbId `form:"min_id"`
	Limit int                   `form:"limit"`
}

func (v *peerView) GetPeers(c *gin.Context) {
	ctx := c.Request.Context()
	user := middleware.GetUser(c.Request.Context())

	request := GetPeerRequest{}
	if err := c.ShouldBind(&request); err != nil {
		utility.AbortBadRequest(c, err)
		return
	}

	pagination := repository_server.PaginationOption[model_server.PeerDbId]{
		MinId: request.MinId,
		Limit: request.Limit,
	}
	peers, err := v.controller.GetPeersByUser(ctx, user, pagination)
	if err != nil {
		utility.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, peers)
	return
}

type CreatePeerRequest struct {
	ProtocolType     string                `form:"protocol_type"`
	VersionMajor     uint32                `form:"version_major"`
	VersionMinor     uint32                `form:"version_minor"`
	ConnectionUri    string                `form:"connection_uri"`
	PeerPemPublicKey *multipart.FileHeader `form:"peer_pem_public_key"`
}

func (v *peerView) CreatePeer(c *gin.Context) {
	ctx := c.Request.Context()
	user := middleware.GetUser(c.Request.Context())

	request := CreatePeerRequest{}
	if err := c.ShouldBind(&request); err != nil {
		utility.AbortBadRequest(c, err)
		return
	}

	publicKey := ""
	{
		file, err := request.PeerPemPublicKey.Open()
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

	peer, err := v.controller.AddPeerToUser(
		ctx,
		user,
		request.ProtocolType,
		request.VersionMajor,
		request.VersionMinor,
		request.ConnectionUri,
		publicKey,
	)
	if err != nil {
		utility.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, peer)
	return
}
