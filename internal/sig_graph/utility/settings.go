package utility_sig_graph

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"sig_graph_scp/pkg/utility"
	"strings"
)

type settings struct {
	peerAddresses       []string
	channelName         string
	contractName        string
	mspId               string
	x509CertificateData string
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

		base64Decoded, err := base64.StdEncoding.DecodeString(string(fileData))
		ret.x509CertificateData = string(base64Decoded)
	}

	return ret, nil
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

func (s *settings) X509CertificateData() string {
	return s.x509CertificateData
}
