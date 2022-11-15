package model_sig_graph

import (
	"sig_graph_scp/pkg/model"

	"github.com/shopspring/decimal"
)

type Asset struct {
	Node            `mapstructure:",squash"`
	CreationProcess model.ECreationProcess `json:"creation_process" mapstructure:"creation_process"`
	Unit            string                 `json:"unit"  mapstructure:"unit"`
	Quantity        decimal.Decimal        `json:"quantity"  mapstructure:"quantity"`
	MaterialName    string                 `json:"material_name"  mapstructure:"material_name"`
}

func NewAsset(
	node Node,
	creationProcess model.ECreationProcess,
	unit string,
	quantity decimal.Decimal,
	materialName string,
) Asset {
	return Asset{
		Node:            node,
		CreationProcess: creationProcess,
		Unit:            unit,
		Quantity:        quantity,
		MaterialName:    materialName,
	}
}
