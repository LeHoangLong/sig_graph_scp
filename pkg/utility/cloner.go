package utility

import (
	"context"
	"encoding/json"
)

type cloner struct {
}

func NewCloner() *cloner {
	return &cloner{}
}

func (c *cloner) Clone(ctx context.Context, src any, dst any) error {
	srcStr, err := json.Marshal(src)
	if err != nil {
		return err
	}

	err = json.Unmarshal(srcStr, dst)
	if err != nil {
		return err
	}
	return nil
}
