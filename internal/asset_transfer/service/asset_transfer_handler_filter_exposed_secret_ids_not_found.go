package service_asset_transfer

import (
	"context"
	model_asset_transfer "sig_graph_scp/pkg/asset_transfer/model"
	api_sig_graph "sig_graph_scp/pkg/sig_graph/api"
	"sig_graph_scp/pkg/utility"
	"time"
)

type assetTransferHandlerFilterExposedSecretIdsNotFound struct {
	handler     AssetTransferHandlerI
	sigGraphApi api_sig_graph.SigGraphClientApi
}

func NewAssetTransferHandlerFilterExposedSecretIdsNotFound(
	handler AssetTransferHandlerI,
	sigGraphApi api_sig_graph.SigGraphClientApi,
) *assetTransferHandlerFilterExposedSecretIdsNotFound {
	return &assetTransferHandlerFilterExposedSecretIdsNotFound{
		handler:     handler,
		sigGraphApi: sigGraphApi,
	}
}

func (s *assetTransferHandlerFilterExposedSecretIdsNotFound) HandleAssetTransfer(
	ctx context.Context,
	ackId string,
	requestTime *time.Time,
	assetId string,
	senderPublicKey string,
	recipientPublicKey string,
	exposedSecretIds map[string]model_asset_transfer.PrivateId,
	candidates []model_asset_transfer.CandidateId,
) error {
	// verify that nodes exist
	ids := map[string]bool{}
	ids[assetId] = true
	for hash := range exposedSecretIds {
		thisId := exposedSecretIds[hash].ThisId
		otherId := exposedSecretIds[hash].OtherId
		ids[string(thisId)] = true
		ids[string(otherId)] = true
	}
	exists, err := s.sigGraphApi.DoNodeIdsExists(ctx, ids)
	if err != nil {
		return err
	}

	// return error if asset not found
	if !exists[assetId] {
		return utility.ErrNotFound
	}

	foundExposedSecretIds := map[string]model_asset_transfer.PrivateId{}
	for hash := range exposedSecretIds {
		thisId := exposedSecretIds[hash].ThisId
		otherId := exposedSecretIds[hash].OtherId
		if exists[string(thisId)] && exists[string(otherId)] {
			foundExposedSecretIds[hash] = exposedSecretIds[hash]
		}
	}

	return s.handler.HandleAssetTransfer(
		ctx,
		ackId,
		requestTime,
		assetId,
		senderPublicKey,
		recipientPublicKey,
		foundExposedSecretIds,
		candidates,
	)
}
