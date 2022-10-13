package model

import (
	"github.com/shopspring/decimal"
)

type ECreationProcess = string

const (
	ECreationProcessCreate   = "create"
	ECreationProcessTransfer = "transfer"
)

type Asset struct {
	Node
	CreationProcess ECreationProcess `json:"creation_process"`
	Unit            string           `json:"unit"`
	Quantity        decimal.Decimal  `json:"quantity"`
	MaterialName    string           `json:"material_name"`
}
