package api_asset_transfer

import (
	service_asset_transfer "sig_graph_scp/internal/asset_transfer/service"
	api_sig_graph "sig_graph_scp/pkg/sig_graph/api"

	EventBus "github.com/asaskevich/eventbus"
)

func NewAssetTransferHandlerEventBus(bus EventBus.Bus, topicName string) (AssetTransferHandlerI, error) {
	return service_asset_transfer.NewAssetTransferHandlerEventBus(bus, topicName), nil
}

func NewAssetTransferHandlerFilterExposedSecretIdsInvalidHash(
	handler AssetTransferHandlerI,
) (AssetTransferHandlerI, error) {
	return service_asset_transfer.NewAssetTransferHandlerFilterExposedSecretIdsInvalidHash(
		handler,
	), nil
}

func NewAssetTransferHandlerFilterExposedSecretIdsNotFound(
	handler AssetTransferHandlerI,
	sigGraphClient api_sig_graph.SigGraphClientApi,
) (AssetTransferHandlerI, error) {
	return service_asset_transfer.NewAssetTransferHandlerFilterExposedSecretIdsNotFound(
		handler,
		sigGraphClient,
	), nil
}
