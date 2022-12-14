package service_asset_transfer

import (
	"context"
	model_asset_transfer "sig_graph_scp/pkg/asset_transfer/model"

	EventBus "github.com/asaskevich/eventbus"
)

type assetAcceptHandlerEventBus struct {
	bus       EventBus.Bus
	topicName string
}

func NewAssetAcceptHandlerEventBus(
	bus EventBus.Bus,
	topicName string,
) *assetAcceptHandlerEventBus {
	return &assetAcceptHandlerEventBus{
		bus:       bus,
		topicName: topicName,
	}
}

func (s *assetAcceptHandlerEventBus) HandleAssetAccept(
	ctx context.Context,
	ackId string,
	isAcceptedOrRejected bool,
	message string,
) {
	event := model_asset_transfer.AcceptAssetEvent{
		AckId:      ackId,
		IsAccepted: isAcceptedOrRejected,
		Message:    message,
	}
	s.bus.Publish(s.topicName, event)
}
