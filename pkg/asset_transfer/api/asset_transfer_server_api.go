package api_asset_transfer

import (
	"context"
	service_asset_transfer "sig_graph_scp/internal/asset_transfer/service"
	api_sig_graph "sig_graph_scp/pkg/sig_graph/api"

	EventBus "github.com/asaskevich/eventbus"
)

type AssetTransferServerApi interface {
	// will block, so you should call this function inside a goroutine
	Start() error
}

type AssetTransferHandlerI interface {
	service_asset_transfer.AssetTransferHandlerI
}

type AssetTransferServerApiOptions struct {
	CustomHandlers    []AssetTransferHandlerI
	BusName           string
	SigGraphApiClient api_sig_graph.SigGraphClientApi
}

type assetTransferServerApi struct {
	assetTransferServer service_asset_transfer.AssetTransferServerI
}

func NewAssetTransferServerApi(
	serverAddress string,
	option *AssetTransferServerApiOptions,
) (AssetTransferServerApi, error) {
	multiAssetTransferHandler := service_asset_transfer.NewAssetTransferHandlerMultiple([]service_asset_transfer.AssetTransferHandlerI{})
	ctx := context.Background()
	if option != nil && option.CustomHandlers != nil {
		for i := range option.CustomHandlers {
			err := multiAssetTransferHandler.AddHandler(ctx, option.CustomHandlers[i])
			if err != nil {
				return nil, err
			}
		}
	} else {
		busName := "new_request_to_accept_asset_event"
		if option != nil && option.BusName != "" {
			busName = option.BusName
		}
		bus := EventBus.New()

		eventBus, err := NewAssetTransferHandlerEventBus(bus, busName)
		if err != nil {
			return nil, err
		}

		secretIdFilterInvalidHash, err := NewAssetTransferHandlerFilterExposedSecretIdsInvalidHash(eventBus)
		if err != nil {
			return nil, err
		}

		if option != nil && option.SigGraphApiClient != nil {
			secretIdFilterNotFound, err := NewAssetTransferHandlerFilterExposedSecretIdsNotFound(
				secretIdFilterInvalidHash,
				option.SigGraphApiClient,
			)

			if err != nil {
				return nil, err
			}

			err = multiAssetTransferHandler.AddHandler(ctx, secretIdFilterNotFound)
			if err != nil {
				return nil, err
			}
		} else {
			err = multiAssetTransferHandler.AddHandler(ctx, secretIdFilterInvalidHash)
			if err != nil {
				return nil, err
			}
		}
	}

	assetTransferServer := service_asset_transfer.NewAssetTransferServerGrpc(
		multiAssetTransferHandler,
		serverAddress,
	)
	return &assetTransferServerApi{
		assetTransferServer: assetTransferServer,
	}, nil
}

func (a *assetTransferServerApi) Start() error {
	return a.assetTransferServer.Start()
}
