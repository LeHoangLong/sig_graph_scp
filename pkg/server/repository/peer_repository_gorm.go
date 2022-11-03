package repository_server

import (
	"context"
	"fmt"
	model_server "sig_graph_scp/pkg/server/model"
	"sig_graph_scp/pkg/utility"

	"gorm.io/gorm"
)

type peerRepositoryGorm struct {
	transactionManagerGorm *transactionManagerGorm
}

func NewPeerRepositoryGorm(
	transactionManagerGorm *transactionManagerGorm,
) *peerRepositoryGorm {
	return &peerRepositoryGorm{
		transactionManagerGorm: transactionManagerGorm,
	}
}

type gormPeerProtocol struct {
	ID           uint32
	ProtocolType string
	VersionMajor uint32
	VersionMinor uint32
}

type gormPeer struct {
	ID               uint64
	UserId           uint64
	ProtocolId       uint32
	ConnectionUri    string
	PeerPemPublicKey string
}

type gormFetchPeer struct {
	ID               uint64
	UserId           uint64
	ProtocolType     string
	VersionMajor     uint32
	VersionMinor     uint32
	ConnectionUri    string
	PeerPemPublicKey string
}

func toModelServerPeer(peer *gormFetchPeer) model_server.Peer {
	return model_server.Peer{
		PeerDbId: peer.ID,
		UserId:   peer.UserId,
		Protocol: model_server.PeerProtocol{
			Type:         peer.ProtocolType,
			VersionMajor: peer.VersionMajor,
			VersionMinor: peer.VersionMinor,
		},
		ConnectionUri:    peer.ConnectionUri,
		PeerPemPublicKey: peer.PeerPemPublicKey,
	}
}

func (r *peerRepositoryGorm) FetchPeerById(
	ctx context.Context,
	txId TransactionId,
	id model_server.PeerDbId,
) (*model_server.Peer, error) {
	tx, err := r.transactionManagerGorm.GetTransaction(ctx, txId)
	if err != nil {
		return nil, err
	}

	peer := gormFetchPeer{}

	err = tx.Raw(`
		SELECT
			peer.id, 
			peer.user_id, 
			protocol.protocol_type,
			protocol.version_major,
			protocol.version_minor,
			peer.connection_uri,
			peer.peer_pem_public_key
		FROM gorm_peers peer
			JOIN gorm_peer_protocols protocol 
			ON peer.protocol_id  = protocol.id 
		WHERE peer.id = ?
	`, id).First(&peer).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utility.ErrNotFound
		}

		return nil, err
	}

	modelPeer := toModelServerPeer(&peer)
	return &modelPeer, nil
}

func (r *peerRepositoryGorm) FetchPeersByUser(
	ctx context.Context,
	txId TransactionId,
	user *model_server.User,
	pagination PaginationOption[model_server.PeerDbId],
) ([]model_server.Peer, error) {
	tx, err := r.transactionManagerGorm.GetTransaction(ctx, txId)
	if err != nil {
		return nil, err
	}

	peers := []gormFetchPeer{}

	err = tx.Raw(`
		SELECT
			peer.id, 
			peer.user_id, 
			protocol.protocol_type,
			protocol.version_major,
			protocol.version_minor,
			peer.connection_uri,
			peer.peer_pem_public_key
		FROM gorm_peers peer
			JOIN gorm_peer_protocols protocol 
			ON peer.protocol_id  = protocol.id 
		WHERE peer.user_id = ? AND peer.id >= ?
		ORDER BY peer.id ASC
		LIMIT ?
	`, user.ID, pagination.MinId, pagination.Limit).Find(&peers).Error

	if err != nil {
		return nil, err
	}

	modelPeers := make([]model_server.Peer, 0, len(peers))
	for i := range peers {
		modelPeers = append(modelPeers, toModelServerPeer(&peers[i]))
	}
	return modelPeers, nil
}

func (r *peerRepositoryGorm) AddPeerToUser(
	ctx context.Context,
	txId TransactionId,
	peer *model_server.Peer,
) error {
	tx, err := r.transactionManagerGorm.GetTransaction(ctx, txId)
	if err != nil {
		return err
	}

	protocol := gormPeerProtocol{}
	err = tx.Where(`
		protocol_type = ? 
		AND version_major = ? 
		AND version_minor = ?
	`, peer.Protocol.Type, peer.Protocol.VersionMajor, peer.Protocol.VersionMinor).First(&protocol).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("%w: protocol not supported", utility.ErrNotFound)
		}

		return err
	}

	gormPeer := gormPeer{
		UserId:           peer.UserId,
		ProtocolId:       protocol.ID,
		ConnectionUri:    peer.ConnectionUri,
		PeerPemPublicKey: peer.PeerPemPublicKey,
	}

	err = tx.Create(&gormPeer).Error
	if err != nil {
		return err
	}

	peer.PeerDbId = gormPeer.ID
	return nil
}
