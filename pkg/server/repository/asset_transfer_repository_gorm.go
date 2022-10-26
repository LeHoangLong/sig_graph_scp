package repository_server

import (
	"context"
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

type gormRequestToAcceptAsset struct {
	ID                        model_server.RequestId            `gorm:"primaryKey"`
	Status                    model.ERequestToAcceptAssetStatus `gorm:"column:request_status"`
	IsOutboundOrInbound       bool
	Time                      uint64 `gorm:"column:request_time_ms"`
	AckId                     string
	Accepted                  bool
	AssetId                   model_server.NodeDbId
	PeerId                    model_server.PeerDbId
	UserId                    model_server.UserId
	ExposedPrivateConnections []gormRequestToAcceptAssetExposedPrivateId `gorm:"foreignKey:RequestId"`
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

	gormRequest := gormRequestToAcceptAsset{
		ID:                        request.Id,
		Status:                    request.Status,
		IsOutboundOrInbound:       request.IsOutboundOrInbound,
		Time:                      request.Time,
		AckId:                     request.AckId,
		Accepted:                  request.Accepted,
		AssetId:                   request.AssetId,
		PeerId:                    request.PeerId,
		UserId:                    request.UserId,
		ExposedPrivateConnections: []gormRequestToAcceptAssetExposedPrivateId{},
	}

	for hash := range request.ExposedPrivateConnections {
		gormRequest.ExposedPrivateConnections = append(
			gormRequest.ExposedPrivateConnections,
			gormRequestToAcceptAssetExposedPrivateId{
				PrivateId: request.ExposedPrivateConnections[hash],
			},
		)
	}

	err = tx.Create(&gormRequest).Error
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
	pagination PaginationOption[model_server.RequestId],
) ([]model_server.RequestToAcceptAsset, error) {
	tx, err := r.transactionManager.GetTransaction(ctx, txId)
	if err != nil {
		return nil, err
	}

	gormRequests := []gormRequestToAcceptAsset{}
	err = tx.Where("user_id = ? AND request_status = ? AND id >= ?", user.ID, status, pagination.MinId).
		Limit(pagination.Limit).
		Order("id asc").
		Find(&gormRequests).Error
	if err != nil {
		return nil, err
	}

	requests := make([]model_server.RequestToAcceptAsset, 0, len(gormRequests))
	for i := range gormRequests {
		modelRequest := model_server.RequestToAcceptAsset{
			Id:                  gormRequests[i].ID,
			Status:              gormRequests[i].Status,
			IsOutboundOrInbound: gormRequests[i].IsOutboundOrInbound,
			Time:                gormRequests[i].Time,
			AckId:               gormRequests[i].AckId,
			Accepted:            gormRequests[i].Accepted,
			AssetId:             gormRequests[i].AssetId,
			PeerId:              gormRequests[i].PeerId,
			UserId:              gormRequests[i].UserId,
		}

		for j := range gormRequests[i].ExposedPrivateConnections {
			hash := gormRequests[i].ExposedPrivateConnections[j].PrivateId.ThisHash
			modelRequest.ExposedPrivateConnections[hash] = gormRequests[i].ExposedPrivateConnections[j].PrivateId
		}

		requests = append(requests, modelRequest)
	}

	return requests, nil
}
