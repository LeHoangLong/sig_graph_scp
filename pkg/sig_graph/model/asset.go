package model_sig_graph

import (
	"sig_graph_scp/pkg/model"

	"github.com/shopspring/decimal"
)

type Asset struct {
	Node
	CreationProcess model.ECreationProcess `json:"creation_process"`
	Unit            string                 `json:"unit"`
	Quantity        decimal.Decimal        `json:"quantity"`
	MaterialName    string                 `json:"material_name"`
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
