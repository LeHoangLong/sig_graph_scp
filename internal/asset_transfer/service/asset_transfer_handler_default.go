package service_asset_transfer

import (
	"context"
	model_asset_transfer "sig_graph_scp/pkg/asset_transfer/model"
	"time"
)

type assetTransferHandlerDefault struct {
}

func NewAssetTransferHandlerDefault() *assetTransferHandlerDefault {
	return &assetTransferHandlerDefault{}
}

func (s *assetTransferHandlerDefault) HandleAssetTransfer(
	ctx context.Context,
	ackId string,
	requestTime *time.Time,
	assetId string,
	senderPublicKey string,
	recipientPublicKey string,
	exposedSecretIds map[string]model_asset_transfer.PrivateId,
	candidates []model_asset_transfer.CandidateId,
) error {
	return nil
}
