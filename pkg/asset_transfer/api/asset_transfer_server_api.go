package api_asset_transfer

import (
	"context"
	service_asset_transfer "sig_graph_scp/internal/asset_transfer/service"
	api_sig_graph "sig_graph_scp/pkg/sig_graph/api"
	"sig_graph_scp/pkg/utility"

	EventBus "github.com/asaskevich/eventbus"
)

type AssetTransferServerApi interface {
	// will block, so you should call this function inside a goroutine
	Start() error
	GetDefaultNewReceivedRequestToAcceptAssetTopic() string
	GetDefaultNewReceivedAssetAcceptTopic() string
}

type AssetTransferHandlerI interface {
	service_asset_transfer.AssetTransferHandlerI
}

type AssetAcceptHandlerI interface {
	service_asset_transfer.AssetAcceptHandlerI
}

type AssetTransferServerApiOptions struct {
	CustomHandlers                       []AssetTransferHandlerI
	NewReceivedRequestToAcceptAssetTopic string
	NewReceivedAssetAcceptTopic          string
	SigGraphApiClient                    api_sig_graph.SigGraphClientApi
	EventBus                             EventBus.Bus
}

type assetTransferServerApi struct {
	assetTransferServer                 service_asset_transfer.AssetTransferServerI
	newReceivedRequesToAcceptAssetTopic string
	newAssetAcceptTopic                 string
}

const defaultNewReceivedRequestToAcceptAssetTopic = "new_request_to_accept_asset_event"
const defaultNewReceivedAssetAcceptTopic = "new_received_asset_accept_topic"

func NewAssetTransferServerApi(
	serverAddress string,
	option AssetTransferServerApiOptions,
) (AssetTransferServerApi, error) {
	multiAssetTransferHandler := service_asset_transfer.NewAssetTransferHandlerMultiple([]service_asset_transfer.AssetTransferHandlerI{})
	ctx := context.Background()
	if option.CustomHandlers != nil {
		for i := range option.CustomHandlers {
			err := multiAssetTransferHandler.AddHandler(ctx, option.CustomHandlers[i])
			if err != nil {
				return nil, err
			}
		}
	} else {
		var handler AssetTransferHandlerI
		handler, err := NewAssetTransferHandlerDefault()
		if err != nil {
			return nil, err
		}

		if option.EventBus != nil {
			newReceivedRequesToAcceptAssetTopic := defaultNewReceivedRequestToAcceptAssetTopic
			if option.NewReceivedRequestToAcceptAssetTopic != "" {
				newReceivedRequesToAcceptAssetTopic = option.NewReceivedRequestToAcceptAssetTopic
			}

			var err error
			handler, err = NewAssetTransferHandlerEventBus(option.EventBus, newReceivedRequesToAcceptAssetTopic)
			if err != nil {
				return nil, err
			}
		}

		secretIdFilterInvalidHash, err := NewAssetTransferHandlerFilterExposedSecretIdsInvalidHash(handler)
		if err != nil {
			return nil, err
		}

		if option.SigGraphApiClient != nil {
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

	assetAcceptHandler, err := NewAssetAcceptHandlerDefault()
	if err != nil {
		return nil, err
	}

	if option.EventBus != nil {
		topicName := defaultNewReceivedAssetAcceptTopic
		if option.NewReceivedAssetAcceptTopic != "" {
			topicName = option.NewReceivedAssetAcceptTopic
		}

		assetAcceptHandler, err = NewAssetAcceptHandlerEventBus(option.EventBus, topicName)
		if err != nil {
			return nil, err
		}
	}

	hashedIdGenerator := utility.NewHashedIdGeneratorService()

	assetTransferServer := service_asset_transfer.NewAssetTransferServerGrpc(
		multiAssetTransferHandler,
		assetAcceptHandler,
		serverAddress,
		hashedIdGenerator,
	)
	return &assetTransferServerApi{
		assetTransferServer: assetTransferServer,
	}, nil
}

func (a *assetTransferServerApi) GetDefaultNewReceivedRequestToAcceptAssetTopic() string {
	return defaultNewReceivedRequestToAcceptAssetTopic
}

func (a *assetTransferServerApi) GetDefaultNewReceivedAssetAcceptTopic() string {
	return defaultNewReceivedAssetAcceptTopic
}

func (a *assetTransferServerApi) Start() error {
	return a.assetTransferServer.Start()
}
