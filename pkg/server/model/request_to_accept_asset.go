package model_server

import (
	model_asset_transfer "sig_graph_scp/pkg/asset_transfer/model"
	"sig_graph_scp/pkg/model"
)

type RequestId uint64

type RequestToAcceptAsset struct {
	Id                        RequestId                         `json:"id"`
	Status                    model.ERequestToAcceptAssetStatus `json:"status"`
	IsOutboundOrInbound       bool                              `json:"is_outbound_or_inbound"`
	Time                      uint64                            `json:"time"`
	AckId                     string                            `json:"ack_id"`
	Accepted                  bool                              `json:"accepted"`
	AssetId                   NodeDbId                          `json:"asset_id"`
	PeerId                    PeerDbId                          `json:"peer_id"`
	UserId                    UserId                            `json:"user_id"`
	ExposedPrivateConnections map[string]PrivateId              `json:"exposed_private_connections"`
}

func FromAssetTransferRequestToAcceptAsset(
	Id RequestId,
	AssetId NodeDbId,
	PeerId PeerDbId,
	UserId UserId,
	request *model_asset_transfer.RequestToAcceptAsset,
) RequestToAcceptAsset {
	exposedSecretIds := map[string]PrivateId{}
	for hash := range request.ExposedPrivateConnections {
		exposedSecretIds[hash] = PrivateId{
			ThisId:     NodeId(request.ExposedPrivateConnections[hash].ThisId),
			ThisSecret: request.ExposedPrivateConnections[hash].ThisSecret,
			ThisHash:   request.ExposedPrivateConnections[hash].ThisHash,

			OtherId:     NodeId(request.ExposedPrivateConnections[hash].OtherId),
			OtherSecret: request.ExposedPrivateConnections[hash].OtherSecret,
			OtherHash:   request.ExposedPrivateConnections[hash].OtherHash,
		}
	}

	return RequestToAcceptAsset{
		Id:                        Id,
		Status:                    request.Status,
		IsOutboundOrInbound:       request.IsOutboundOrInbound,
		Time:                      request.TimeMs,
		AckId:                     request.AckId,
		Accepted:                  request.Accepted,
		AssetId:                   AssetId,
		PeerId:                    PeerId,
		UserId:                    UserId,
		ExposedPrivateConnections: exposedSecretIds,
	}
}
