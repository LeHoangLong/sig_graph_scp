package repository_server

import (
	"context"
	"database/sql"
	"sig_graph_scp/pkg/model"
	model_server "sig_graph_scp/pkg/server/model"
)

type assetTransferRepositoryGorm struct {
	transactionManager transactionManagerGorm
}

func NewAssetTransferRepositoryGorm(
	transactionManager transactionManagerGorm,
) *assetTransferRepositoryGorm {
	return &assetTransferRepositoryGorm{
		transactionManager: transactionManager,
	}
}

type gormRequestToAcceptAssetExposedPrivateId struct {
	ID                     uint64 `gorm:"primaryKey"`
	RequestId              model_server.RequestId
	model_server.PrivateId `gorm:"embedded"`
}

type gormRequestToAcceptAssetCandidateId struct {
	ID          uint64 `gorm:"primaryKey"`
	RequestId   model_server.RequestId
	CandidateId string `gorm:"column:candidate_id"`
	Secret      string `gorm:"column:candidate_secret"`
	Signature   string `gorm:"column:candidate_signature"`
}

type gormRequestToAcceptAsset struct {
	ID                        model_server.RequestId            `gorm:"primaryKey"`
	Status                    model.ERequestToAcceptAssetStatus `gorm:"column:request_status"`
	IsOutboundOrInbound       bool
	Time                      uint64 `gorm:"column:request_time_ms"`
	AckId                     string
	AssetId                   model_server.NodeDbId
	NewAssetId                sql.NullInt64
	PeerId                    model_server.PeerDbId
	UserId                    model_server.UserId
	ExposedPrivateConnections []gormRequestToAcceptAssetExposedPrivateId `gorm:"foreignKey:RequestId"`
	CandidateIds              []gormRequestToAcceptAssetCandidateId      `gorm:"foreignKey:RequestId"`
	AcceptMessage             string
}

func (r *assetTransferRepositoryGorm) CreateAssetAcceptRequest(
	ctx context.Context,
	txId TransactionId,
	request *model_server.RequestToAcceptAsset,
) error {
	tx, err := r.transactionManager.GetTransaction(ctx, txId)
	if err != nil {
		return err
	}
	gormRequest := fromRequestToAcceptAsset(request)
	err = tx.Create(&gormRequest).Error
	if err != nil {
		return err
	}

	request.Id = gormRequest.ID
	return nil
}

func (r *assetTransferRepositoryGorm) UpdateAssetAcceptRequest(
	ctx context.Context,
	txId TransactionId,
	request *model_server.RequestToAcceptAsset,
) error {
	tx, err := r.transactionManager.GetTransaction(ctx, txId)
	if err != nil {
		return err
	}
	gormRequest := fromRequestToAcceptAsset(request)
	err = tx.Save(&gormRequest).Error
	if err != nil {
		return err
	}

	request.Id = gormRequest.ID
	return nil
}

func (r *assetTransferRepositoryGorm) FetchAssetAcceptRequestsByUserAndStatus(
	ctx context.Context,
	txId TransactionId,
	user *model_server.User,
	status model.ERequestToAcceptAssetStatus,
	outboundOrInbound bool,
	pagination PaginationOption[model_server.RequestId],
) ([]model_server.RequestToAcceptAsset, error) {
	tx, err := r.transactionManager.GetTransaction(ctx, txId)
	if err != nil {
		return nil, err
	}

	gormRequests := []gormRequestToAcceptAsset{}
	err = tx.Preload("ExposedPrivateConnections").Preload("CandidateIds").Where("user_id = ? AND request_status = ? AND id >= ? AND is_outbound_or_inbound = ?", user.ID, status, pagination.MinId, outboundOrInbound).
		Limit(pagination.Limit).
		Order("id asc").
		Find(&gormRequests).Error
	if err != nil {
		return nil, err
	}

	requests := make([]model_server.RequestToAcceptAsset, 0, len(gormRequests))
	for i := range gormRequests {
		modelRequest := toModelRequest(&gormRequests[i])
		requests = append(requests, modelRequest)
	}

	return requests, nil
}

func (r *assetTransferRepositoryGorm) FetchAssetAcceptRequestsById(
	ctx context.Context,
	txId TransactionId,
	user *model_server.User,
	id model_server.RequestId,
) (*model_server.RequestToAcceptAsset, error) {
	tx, err := r.transactionManager.GetTransaction(ctx, txId)
	if err != nil {
		return nil, err
	}

	gormRequest := gormRequestToAcceptAsset{}
	err = tx.Preload("ExposedPrivateConnections").Preload("CandidateIds").Where("user_id = ? AND id = ?", user.ID, id).
		First(&gormRequest).Error
	if err != nil {
		return nil, err
	}

	modelRequest := toModelRequest(&gormRequest)
	return &modelRequest, nil
}

func (r *assetTransferRepositoryGorm) FetchAssetAcceptRequestsByAckId(
	ctx context.Context,
	txId TransactionId,
	ackId string,
	outboundOrInbound bool,
) (*model_server.RequestToAcceptAsset, error) {
	tx, err := r.transactionManager.GetTransaction(ctx, txId)
	if err != nil {
		return nil, err
	}

	gormRequest := gormRequestToAcceptAsset{}
	err = tx.Preload("ExposedPrivateConnections").Preload("CandidateIds").Where("ack_id = ? AND is_outbound_or_inbound = ?", ackId, outboundOrInbound).
		First(&gormRequest).Error
	if err != nil {
		return nil, err
	}

	modelRequest := toModelRequest(&gormRequest)
	return &modelRequest, nil

}

func toModelRequest(gormRequest *gormRequestToAcceptAsset) model_server.RequestToAcceptAsset {
	var newAssetId *model_server.NodeDbId = nil
	if gormRequest.NewAssetId.Valid {
		newAssetId = new(model_server.NodeDbId)
		*newAssetId = model_server.NodeDbId(gormRequest.NewAssetId.Int64)
	}

	modelRequest := model_server.RequestToAcceptAsset{
		Id:                        gormRequest.ID,
		Status:                    gormRequest.Status,
		IsOutboundOrInbound:       gormRequest.IsOutboundOrInbound,
		Time:                      gormRequest.Time,
		AckId:                     gormRequest.AckId,
		AssetId:                   gormRequest.AssetId,
		NewAssetId:                newAssetId,
		PeerId:                    gormRequest.PeerId,
		UserId:                    gormRequest.UserId,
		AcceptMessage:             gormRequest.AcceptMessage,
		ExposedPrivateConnections: map[string]model_server.PrivateId{},
		CandidateIds:              []model_server.CandidateId{},
	}

	for j := range gormRequest.ExposedPrivateConnections {
		hash := gormRequest.ExposedPrivateConnections[j].PrivateId.ThisHash
		modelRequest.ExposedPrivateConnections[hash] = gormRequest.ExposedPrivateConnections[j].PrivateId
	}

	for j := range gormRequest.CandidateIds {
		modelRequest.CandidateIds = append(
			modelRequest.CandidateIds,
			model_server.CandidateId{
				Id:        gormRequest.CandidateIds[j].CandidateId,
				Secret:    gormRequest.CandidateIds[j].Secret,
				Signature: gormRequest.CandidateIds[j].Signature,
			},
		)
	}

	return modelRequest
}

func fromRequestToAcceptAsset(
	request *model_server.RequestToAcceptAsset,
) gormRequestToAcceptAsset {
	newAssetId := sql.NullInt64{}
	if request.NewAssetId != nil {
		newAssetId.Valid = true
		newAssetId.Int64 = int64(*request.NewAssetId)
	} else {
		newAssetId.Valid = false
	}

	gormRequest := gormRequestToAcceptAsset{
		ID:                        request.Id,
		Status:                    request.Status,
		IsOutboundOrInbound:       request.IsOutboundOrInbound,
		Time:                      request.Time,
		AckId:                     request.AckId,
		NewAssetId:                newAssetId,
		AssetId:                   request.AssetId,
		PeerId:                    request.PeerId,
		UserId:                    request.UserId,
		AcceptMessage:             request.AcceptMessage,
		ExposedPrivateConnections: []gormRequestToAcceptAssetExposedPrivateId{},
		CandidateIds:              []gormRequestToAcceptAssetCandidateId{},
	}

	for hash := range request.ExposedPrivateConnections {
		gormRequest.ExposedPrivateConnections = append(
			gormRequest.ExposedPrivateConnections,
			gormRequestToAcceptAssetExposedPrivateId{
				PrivateId: request.ExposedPrivateConnections[hash],
			},
		)
	}

	for i := range request.CandidateIds {
		gormRequest.CandidateIds = append(
			gormRequest.CandidateIds,
			gormRequestToAcceptAssetCandidateId{
				CandidateId: request.CandidateIds[i].Id,
				Secret:      request.CandidateIds[i].Secret,
				Signature:   request.CandidateIds[i].Signature,
			},
		)
	}

	return gormRequest
}
