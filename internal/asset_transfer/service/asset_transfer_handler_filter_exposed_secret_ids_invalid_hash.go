package service_asset_transfer

import (
	"context"
	"crypto/sha512"
	"fmt"
	model_asset_transfer "sig_graph_scp/pkg/asset_transfer/model"
	"time"
)

type assetTransferHandlerFilterExposedSecretIdsInvalidHash struct {
	handler AssetTransferHandlerI
}

func NewAssetTransferHandlerFilterExposedSecretIdsInvalidHash(
	handler AssetTransferHandlerI,
) *assetTransferHandlerFilterExposedSecretIdsInvalidHash {
	return &assetTransferHandlerFilterExposedSecretIdsInvalidHash{
		handler: handler,
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
		thisSecretId := fmt.Sprintf("%s%s", thisId, thisSecret)
		thisHashByte := sha512.Sum512([]byte(thisSecretId))
		thisHash := string(thisHashByte[:])

		if hash != thisHash && hash != exposedSecretIds[hash].ThisHash {
			continue
		}

		otherId := exposedSecretIds[hash].OtherId
		otherSecret := exposedSecretIds[hash].OtherSecret
		otherSecretId := fmt.Sprintf("%s%s", otherId, otherSecret)
		otherHashByte := sha512.Sum512([]byte(otherSecretId))
		otherHash := string(otherHashByte[:])

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
