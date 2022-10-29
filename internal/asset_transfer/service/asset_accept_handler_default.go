package service_asset_transfer

import "context"

type assetAcceptHandlerDefault struct {
}

func NewAssetAcceptHandlerDefault() *assetAcceptHandlerDefault {
	return &assetAcceptHandlerDefault{}
}

func (s *assetAcceptHandlerDefault) HandleAssetAccept(
	ctx context.Context,
	ackId string,
	isAcceptedOrRejected bool,
	message string,
	newId string,
	newSecret string,
	oldId string,
	oldSecret string,
) {

}
