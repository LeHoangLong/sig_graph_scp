package controller_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
	repository_server "sig_graph_scp/pkg/server/repository"
)

type peerController struct {
	transactionManager repository_server.TransactionManagerI
	peerRepository     repository_server.PeerRepositoryI
}

func NewPeerController(
	transactionManager repository_server.TransactionManagerI,
	peerRepository repository_server.PeerRepositoryI,
) *peerController {
	return &peerController{
		transactionManager: transactionManager,
		peerRepository:     peerRepository,
	}
}

func (c *peerController) GetPeersByUser(
	ctx context.Context,
	user *model_server.User,
	pagination repository_server.PaginationOption[model_server.PeerDbId],
) ([]model_server.Peer, error) {
	tx, err := c.transactionManager.BypassTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer c.transactionManager.StopBypassedTransaction(ctx, tx)

	peers, err := c.peerRepository.FetchPeersByUser(ctx, tx, user, pagination)
	if err != nil {
		return nil, err
	}

	return peers, nil
}

func (c *peerController) AddPeerToUser(
	ctx context.Context,
	user *model_server.User,
	protocolType string,
	versionMajor uint32,
	versionMinor uint32,
	connectionUri string,
	peerPemPublicKey string,
) (*model_server.Peer, error) {
	tx, err := c.transactionManager.BypassTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer c.transactionManager.StopBypassedTransaction(ctx, tx)

	peer := model_server.Peer{
		UserId: user.ID,
		Protocol: model_server.PeerProtocol{
			Type:         protocolType,
			VersionMajor: versionMajor,
			VersionMinor: versionMinor,
		},
		ConnectionUri:    connectionUri,
		PeerPemPublicKey: peerPemPublicKey,
	}

	err = c.peerRepository.AddPeerToUser(ctx, tx, &peer)
	if err != nil {
		return nil, err
	}

	return &peer, nil
}
