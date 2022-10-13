package service_sig_graph

import (
	"crypto/x509"
	utility_sig_graph "sig_graph_scp/internal/sig_graph/utility"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type smartContractServiceHyperledger struct {
	contract *client.Contract
}

func NewSmartContractServiceHyperledger() (SmartContractServiceI, error) {
	settings := utility_sig_graph.GetGlobalSettings()

	certificate, err := x509.ParseCertificate([]byte(settings.X509CertificateData()))
	if err != nil {
		return nil, err
	}

	clientConnection, err := grpc.Dial(settings.PeerAddresses()[0], grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer clientConnection.Close()

	mspId := settings.MspId()

	id, err := identity.NewX509Identity(mspId, certificate)
	if err != nil {
		return nil, err
	}

	gateway, err := client.Connect(id, client.WithClientConnection(clientConnection))

	network := gateway.GetNetwork(settings.ChannelName())
	contract := network.GetContract(settings.ContractName())

	service := smartContractServiceHyperledger{
		contract: contract,
	}

	return &service, nil
}

func (s *smartContractServiceHyperledger) CreateTransaction(
	functionName string,
	args ...string,
) (string, error) {
	result, err := s.contract.SubmitTransaction(functionName, args...)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

func (s *smartContractServiceHyperledger) Query(
	functionName string,
	args ...string,
) (string, error) {
	result, err := s.contract.EvaluateTransaction(functionName, args...)
	if err != nil {
		return "", err
	}

	return string(result), nil

}
