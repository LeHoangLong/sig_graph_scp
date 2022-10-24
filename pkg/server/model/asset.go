package model_server

import (
	"sig_graph_scp/pkg/model"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"

	"github.com/shopspring/decimal"
)

type Asset struct {
	Node
	CreationProcess model.ECreationProcess `json:"creation_process"`
	Unit            string                 `json:"unit"`
	Quantity        decimal.Decimal        `json:"quantity"`
	MaterialName    string                 `json:"material_name"`
}

func FromSigGraphAsset(asset *model_sig_graph.Asset, modelNode *Node) Asset {
	return Asset{
		Node:            *modelNode,
		CreationProcess: asset.CreationProcess,
		Unit:            asset.Unit,
		Quantity:        asset.Quantity,
		MaterialName:    asset.MaterialName,
	}
}

func ToSigGraphAsset(asset *Asset) model_sig_graph.Asset {
	node := ToSigGraphNode(&asset.Node)
	return model_sig_graph.Asset{
		Node:            node,
		CreationProcess: asset.CreationProcess,
		Unit:            asset.Unit,
		Quantity:        asset.Quantity,
		MaterialName:    asset.MaterialName,
	}
}
