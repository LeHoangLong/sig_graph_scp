package service_sig_graph

import (
	"crypto/x509"
	"fmt"
	utility_sig_graph "sig_graph_scp/internal/sig_graph/utility"
	"sig_graph_scp/pkg/utility"
	"strings"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type smartContractServiceHyperledger struct {
	contract         *client.Contract
	clientConnection *grpc.ClientConn
}

// newGrpcConnection creates a gRPC connection to the Gateway server.
func newGrpcConnection() (*grpc.ClientConn, error) {
	settings := utility_sig_graph.GetGlobalSettings()
	certificate := settings.TlsX509Certificate()
	gatewayPeer := settings.GatewayPeer()
	peerEndpoint := settings.PeerAddresses()[0]

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)

	connection, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		return nil, err
	}

	return connection, nil
}

func newSign() identity.Sign {
	privateKey := utility_sig_graph.GetGlobalSettings().IdentityEDCSAKey()

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	return sign
}

func NewSmartContractServiceHyperledger() (SmartContractServiceI, error) {
	settings := utility_sig_graph.GetGlobalSettings()

	certificate := settings.IdentityX509Certificate()

	clientConnection, err := newGrpcConnection()
	if err != nil {
		return nil, err
	}

	mspId := settings.MspId()

	id, err := identity.NewX509Identity(mspId, certificate)
	if err != nil {
		clientConnection.Close()
		return nil, err
	}

	gateway, err := client.Connect(
		id,
		client.WithSign(newSign()),
		client.WithClientConnection(clientConnection),
	)

	network := gateway.GetNetwork(settings.ChannelName())
	contract := network.GetContract(settings.ContractName())

	service := smartContractServiceHyperledger{
		contract: contract,
	}
	service.clientConnection = clientConnection

	return &service, nil
}

func (s *smartContractServiceHyperledger) Close() {
	s.clientConnection.Close()
}

func wrapError(err error) error {
	statusErr := status.Convert(err)
	for _, detail := range statusErr.Details() {
		switch detail := detail.(type) {
		case *gateway.ErrorDetail:
			fmt.Println("err 1 ", detail)
			if strings.Contains(detail.Message, "not found") {
				err = multierr.Append(err, utility.ErrNotFound)
			}
		}
	}
	return err
}

func (s *smartContractServiceHyperledger) CreateTransaction(
	functionName string,
	args ...string,
) (string, error) {
	result, err := s.contract.SubmitTransaction(functionName, args...)
	if err != nil {
		return "", wrapError(err)
	}

	return string(result), nil
}

func (s *smartContractServiceHyperledger) Query(
	functionName string,
	args ...string,
) (string, error) {
	result, err := s.contract.EvaluateTransaction(functionName, args...)
	if err != nil {
		return "", wrapError(err)
	}

	return string(result), nil

}
