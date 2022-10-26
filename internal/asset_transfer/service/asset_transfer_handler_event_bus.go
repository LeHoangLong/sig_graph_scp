package service_asset_transfer

import (
	"context"
	"fmt"
	model_asset_transfer "sig_graph_scp/pkg/asset_transfer/model"
	"time"

	EventBus "github.com/asaskevich/eventbus"
)

type assetTransferHandlerEventBus struct {
	bus       EventBus.Bus
	topicName string
}

func NewAssetTransferHandlerEventBus(bus EventBus.Bus, topicName string) *assetTransferHandlerEventBus {
	return &assetTransferHandlerEventBus{
		bus:       bus,
		topicName: topicName,
	}
}

func (s *assetTransferHandlerEventBus) HandleAssetTransfer(
	ctx context.Context,
	ackId string,
	requestTime *time.Time,
	assetId string,
	senderPublicKey string,
	recipientPublicKey string,
	exposedSecretIds map[string]model_asset_transfer.PrivateId,
	candidates []model_asset_transfer.CandidateId,
) error {
	fmt.Println("publishing event ", s.topicName)
	request := model_asset_transfer.RequestToAcceptAssetEvent{
		TimeMs:                    uint64(requestTime.Unix()),
		AssetId:                   assetId,
		AckId:                     ackId,
		PeerPemPublicKey:          senderPublicKey,
		UserPemPublicKey:          recipientPublicKey,
		ExposedPrivateConnections: exposedSecretIds,
		Candidates:                candidates,
	}
	s.bus.Publish(s.topicName, request)
	return nil
}
