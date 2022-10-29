package model_server

import (
	model_asset_transfer "sig_graph_scp/pkg/asset_transfer/model"
	"sig_graph_scp/pkg/model"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"
)

type RequestId uint64

type CandidateId struct {
	Id        string `json:"id"`
	Secret    string `json:"secret"`
	Signature string `json:"signature"`
}

type RequestToAcceptAsset struct {
	Id                        RequestId                         `json:"id"`
	Status                    model.ERequestToAcceptAssetStatus `json:"status"`
	IsOutboundOrInbound       bool                              `json:"is_outbound_or_inbound"`
	Time                      uint64                            `json:"time"`
	AckId                     string                            `json:"ack_id"`
	AssetId                   NodeDbId                          `json:"asset_id"`
	PeerId                    PeerDbId                          `json:"peer_id"`
	UserId                    UserId                            `json:"user_id"`
	ExposedPrivateConnections map[string]PrivateId              `json:"exposed_private_connections"`
	CandidateIds              []CandidateId                     `json:"candidate_ids"`
	AcceptMessage             string                            `json:"accept_message"`
}

func ToAssetTransferRequestToAcceptAsset(
	asset *model_sig_graph.Asset,
	peerPemPublicKey string,
	userKeyPair model_sig_graph.UserKeyPair,
	request *RequestToAcceptAsset,
) model_asset_transfer.RequestToAcceptAsset {
	privateConnections := map[string]model_asset_transfer.PrivateId{}
	for hash := range request.ExposedPrivateConnections {
		privateConnections[hash] = model_asset_transfer.PrivateId{
			ThisId:     string(request.ExposedPrivateConnections[hash].ThisId),
			ThisSecret: request.ExposedPrivateConnections[hash].ThisSecret,
			ThisHash:   request.ExposedPrivateConnections[hash].ThisHash,

			OtherId:     string(request.ExposedPrivateConnections[hash].OtherId),
			OtherSecret: request.ExposedPrivateConnections[hash].OtherSecret,
			OtherHash:   request.ExposedPrivateConnections[hash].OtherHash,
		}
	}

	candidates := make([]model_asset_transfer.CandidateId, 0, len(request.CandidateIds))
	for i := range request.CandidateIds {
		candidates = append(candidates, model_asset_transfer.CandidateId{
			Id:        request.CandidateIds[i].Id,
			Secret:    request.CandidateIds[i].Secret,
			Signature: request.CandidateIds[i].Signature,
		})
	}

	return model_asset_transfer.RequestToAcceptAsset{
		Status:                    request.Status,
		IsOutboundOrInbound:       request.IsOutboundOrInbound,
		TimeMs:                    request.Time,
		AckId:                     request.AckId,
		Asset:                     *asset,
		PeerPemPublicKey:          peerPemPublicKey,
		UserKeyPair:               userKeyPair,
		ExposedPrivateConnections: privateConnections,
		Candidates:                candidates,
	}
}

func FromAssetTransferRequestToAcceptAsset(
	Id RequestId,
	AssetId NodeDbId,
	PeerId PeerDbId,
	UserId UserId,
	AcceptMessage string,
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

	candidateIds := make([]CandidateId, 0, len(request.Candidates))
	for i := range request.Candidates {
		candidateIds = append(candidateIds, CandidateId{
			Id:        request.Candidates[i].Id,
			Secret:    request.Candidates[i].Secret,
			Signature: request.Candidates[i].Signature,
		})
	}

	return RequestToAcceptAsset{
		Id:                        Id,
		Status:                    request.Status,
		IsOutboundOrInbound:       request.IsOutboundOrInbound,
		Time:                      request.TimeMs,
		AckId:                     request.AckId,
		AssetId:                   AssetId,
		PeerId:                    PeerId,
		UserId:                    UserId,
		ExposedPrivateConnections: exposedSecretIds,
		CandidateIds:              candidateIds,
		AcceptMessage:             AcceptMessage,
	}
}
