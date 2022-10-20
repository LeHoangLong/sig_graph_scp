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

func ToModelAsset(asset *Asset, modelNode *model.Node) model.Asset {
	return model.Asset{
		Node:            *modelNode,
		CreationProcess: asset.CreationProcess,
		Unit:            asset.Unit,
		Quantity:        asset.Quantity,
		MaterialName:    asset.MaterialName,
	}
}

func FromModelAsset(asset *model.Asset) Asset {
	node := FromModelNode(&asset.Node)
	return Asset{
		Node:            node,
		CreationProcess: asset.CreationProcess,
		Unit:            asset.Unit,
		Quantity:        asset.Quantity,
		MaterialName:    asset.MaterialName,
	}
}
