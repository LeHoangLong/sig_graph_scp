package service_asset_transfer

import (
	"context"
	model_asset_transfer "sig_graph_scp/pkg/asset_transfer/model"
	"time"
)

// will invoke handlers in reverse order (handlers added later will be called first)
type assetTransferHandlerMultiple struct {
	handlers []AssetTransferHandlerI
}

func NewAssetTransferHandlerMultiple(handlers []AssetTransferHandlerI) *assetTransferHandlerMultiple {
	if handlers == nil {
		handlers = []AssetTransferHandlerI{}
	}

	return &assetTransferHandlerMultiple{
		handlers: handlers,
	}
}

func (s *assetTransferHandlerMultiple) AddHandler(ctx context.Context, handler AssetTransferHandlerI) error {
	s.handlers = append(s.handlers, handler)
	return nil
}

func (s *assetTransferHandlerMultiple) HandleAssetTransfer(
	ctx context.Context,
	ackId string,
	requestTime *time.Time,
	assetId string,
	senderPublicKey string,
	recipientPublicKey string,
	exposedSecretIds map[string]model_asset_transfer.PrivateId,
	candidates []model_asset_transfer.CandidateId,
) error {
	for i := len(s.handlers) - 1; i >= 0; i-- {
		err := s.handlers[i].HandleAssetTransfer(
			ctx,
			ackId,
			requestTime,
			assetId,
			senderPublicKey,
			recipientPublicKey,
			exposedSecretIds,
			candidates,
		)

		if err != nil {
			return err
		}
	}
	return nil
}
