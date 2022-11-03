package utility

import "context"

type HashedIdGeneratorServiceI interface {
	// generate hash id in base64 format
	GenerateHashedId(ctx context.Context, id string, secret string) (string, error)
}
