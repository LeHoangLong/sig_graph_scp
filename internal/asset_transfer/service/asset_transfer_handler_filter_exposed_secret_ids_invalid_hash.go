package service_asset_transfer

import (
	"context"
	model_asset_transfer "sig_graph_scp/pkg/asset_transfer/model"
	"sig_graph_scp/pkg/utility"
	"time"
)

type assetTransferHandlerFilterExposedSecretIdsInvalidHash struct {
	handler       AssetTransferHandlerI
	hashGenerator utility.HashedIdGeneratorServiceI
}

func NewAssetTransferHandlerFilterExposedSecretIdsInvalidHash(
	handler AssetTransferHandlerI,
	hashGenerator utility.HashedIdGeneratorServiceI,
) *assetTransferHandlerFilterExposedSecretIdsInvalidHash {
	return &assetTransferHandlerFilterExposedSecretIdsInvalidHash{
		handler:       handler,
		hashGenerator: hashGenerator,
	}
}

func (s *assetTransferHandlerFilterExposedSecretIdsInvalidHash) HandleAssetTransfer(
	ctx context.Context,
	ackId string,
	requestTime *time.Time,
	assetId string,
	senderPublicKey string,
	recipientPublicKey string,
	exposedSecretIds map[string]model_asset_transfer.PrivateId,
	candidates []model_asset_transfer.CandidateId,
) error {
	passedExposedSecretIds := map[string]model_asset_transfer.PrivateId{}
	for hash := range exposedSecretIds {
		thisId := exposedSecretIds[hash].ThisId
		thisSecret := exposedSecretIds[hash].ThisSecret
		thisHash, err := s.hashGenerator.GenerateHashedId(ctx, thisId, thisSecret)
		if err != nil {
			return err
		}

		if hash != thisHash && hash != exposedSecretIds[hash].ThisHash {
			continue
		}

		otherId := exposedSecretIds[hash].OtherId
		otherSecret := exposedSecretIds[hash].OtherSecret
		otherHash, err := s.hashGenerator.GenerateHashedId(ctx, otherId, otherSecret)
		if err != nil {
			return err
		}

		if otherHash != exposedSecretIds[hash].OtherHash {
			continue
		}

		passedExposedSecretIds[hash] = exposedSecretIds[hash]
	}

	return s.handler.HandleAssetTransfer(
		ctx,
		ackId,
		requestTime,
		assetId,
		senderPublicKey,
		recipientPublicKey,
		passedExposedSecretIds,
		candidates,
	)
}
