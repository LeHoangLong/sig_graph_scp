package service_asset_transfer

import "context"

type AssetAcceptHandlerI interface {
	HandleAssetAccept(
		ctx context.Context,
		ackId string,
		isAcceptedOrRejected bool,
		message string,
	)
}
