package service_asset_transfer

import "context"

type AssetTransferServerI interface {
	RegisterHandler(ctx context.Context, handler AssetTransferHandlerI) error
	RegisterAssetAcceptHandler(ctx context.Context, handler AssetAcceptHandlerI) error
	Start() error
}
