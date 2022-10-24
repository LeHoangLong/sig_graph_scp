package service_asset_transfer

import "context"

type SecretIdGeneratorI interface {
	NewSecretId(ctx context.Context) (string, error)
}
