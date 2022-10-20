package utility_sig_graph

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"sig_graph_scp/pkg/utility"
	"strings"
)

type settings struct {
	peerAddresses           []string
	channelName             string
	contractName            string
	mspId                   string
	identityX509Certificate *x509.Certificate
	identityPeerPrivateKey  *ecdsa.PrivateKey

	TlsCetificate *x509.Certificate
	gatewayPeer   string
}

func NewSettingsFromEnv() (SettingsI, error) {
	ret := &settings{}

	if addresses, ok := os.LookupEnv("PEER_ADDRESSES"); !ok {
		return nil, fmt.Errorf("%w: missing env var PEER_ADDRESSES", utility.ErrNotFound)
	} else {
		addresses := strings.Split(addresses, ",")
		if len(addresses) == 1 {
			return nil, fmt.Errorf("%w: PEER_ADDRESSES must be a comma-seperated list", utility.ErrInvalidArgument)
		}

		ret.peerAddresses = addresses
	}

	if value, ok := os.LookupEnv("CHANNEL_NAME"); !ok {
		return nil, fmt.Errorf("%w: missing env var CHANNEL_NAME", utility.ErrNotFound)
	} else {
		ret.channelName = value
	}

	if value, ok := os.LookupEnv("CONTRACT_NAME"); !ok {
		return nil, fmt.Errorf("%w: missing env var CONTRACT_NAME", utility.ErrNotFound)
	} else {
		ret.contractName = value
	}

	if value, ok := os.LookupEnv("MSP_ID"); !ok {
		return nil, fmt.Errorf("%w: missing env var MSP_ID", utility.ErrNotFound)
	} else {
		ret.mspId = value
	}

	if value, ok := os.LookupEnv("GATEWAY_PEER"); !ok {
		return nil, fmt.Errorf("%w: missing env var GATEWAY_PEER", utility.ErrNotFound)
	} else {
		ret.gatewayPeer = value
	}

	if value, ok := os.LookupEnv("TLS_PEM_CERTIFICATE_PATH"); !ok {
		return nil, fmt.Errorf("%w: missing env var TLS_PEM_CERTIFICATE_PATH", utility.ErrNotFound)
	} else {
		if _, err := os.Stat(value); errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%w: file not found %s", utility.ErrNotFound, value)
		}

		fileData, err := os.ReadFile(value)
		if err != nil {
			return nil, err
		}

		pemDecoded, _ := pem.Decode(fileData)
		if pemDecoded == nil {
			return nil, fmt.Errorf("could not decode pem file")
		}

		certificate, err := x509.ParseCertificate(pemDecoded.Bytes)
		if err != nil {
			return nil, err
		}
		ret.TlsCetificate = certificate
	}

	if value, ok := os.LookupEnv("PEM_CERTIFICATE_PATH"); !ok {
		return nil, fmt.Errorf("%w: missing env var PEM_CERTIFICATE_PATH", utility.ErrNotFound)
	} else {
		if _, err := os.Stat(value); errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%w: file not found %s", utility.ErrNotFound, value)
		}

		fileData, err := os.ReadFile(value)
		if err != nil {
			return nil, err
		}

		pemDecoded, _ := pem.Decode(fileData)
		if pemDecoded == nil {
			return nil, fmt.Errorf("could not decode pem file")
		}

		ret.identityX509Certificate, err = x509.ParseCertificate(pemDecoded.Bytes)
		if err != nil {
			return nil, err
		}
	}

	if value, ok := os.LookupEnv("PEM_SECRET_KEY_PATH"); !ok {
		return nil, fmt.Errorf("%w: missing env var PEM_SECRET_KEY_PATH", utility.ErrNotFound)
	} else {
		if _, err := os.Stat(value); errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%w: file not found %s", utility.ErrNotFound, value)
		}

		fileData, err := os.ReadFile(value)
		if err != nil {
			return nil, err
		}

		pemDecoded, _ := pem.Decode(fileData)
		if pemDecoded == nil {
			return nil, fmt.Errorf("could not decode pem file")
		}

		privateKey, err := x509.ParsePKCS8PrivateKey(pemDecoded.Bytes)
		if err != nil {
			return nil, err
		}
		if privateKey, ok := privateKey.(*ecdsa.PrivateKey); !ok {
			return nil, fmt.Errorf("only ecdsa key supported")
		} else {
			ret.identityPeerPrivateKey = privateKey
		}
	}

	return ret, nil
}

func (s *settings) GatewayPeer() string {
	return s.gatewayPeer
}

func (s *settings) TlsX509Certificate() *x509.Certificate {
	return s.TlsCetificate
}

func (s *settings) PeerAddresses() []string {
	return s.peerAddresses
}

func (s *settings) ChannelName() string {
	return s.channelName
}

func (s *settings) ContractName() string {
	return s.contractName
}

func (s *settings) MspId() string {
	return s.mspId
}

func (s *settings) IdentityX509Certificate() *x509.Certificate {
	return s.identityX509Certificate
}

func (s *settings) IdentityEDCSAKey() *ecdsa.PrivateKey {
	return s.identityPeerPrivateKey
}
