package model

type ECreationProcess = string

const (
	ECreationProcessCreate   = "create"
	ECreationProcessTransfer = "transfer"
)

type ENodeType = string

const (
	ENodeTypeAsset ENodeType = "asset"
)

type EPeerProtocol = string

const (
	EPeerProtocolGrpc EPeerProtocol = "grpc"
)

type ERequestToAcceptAssetStatus = string

const (
	ERequestToAcceptAssetStatusPending  ERequestToAcceptAssetStatus = "pending"
	ERequestToAcceptAssetStatusAccepted ERequestToAcceptAssetStatus = "accepted"
	ERequestToAcceptAssetStatusRejected ERequestToAcceptAssetStatus = "rejected"
)
