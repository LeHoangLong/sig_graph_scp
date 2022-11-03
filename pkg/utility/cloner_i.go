package utility

import "context"

type ClonerI interface {
	Clone(ctx context.Context, src any, dst any) error
}
