package service_asset_transfer

import (
	"context"
	model_asset_transfer "sig_graph_scp/pkg/asset_transfer/model"
)

type CandidateSelectorI interface {
	SelectCandidate(ctx context.Context, candidates []model_asset_transfer.CandidateId) (model_asset_transfer.CandidateId, error)
}
