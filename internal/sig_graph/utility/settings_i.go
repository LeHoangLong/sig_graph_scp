package utility_sig_graph

import (
	"crypto/ecdsa"
	"crypto/x509"
	"os"
	"sync"
)

type SettingsI interface {
	PeerAddresses() []string
	ChannelName() string
	ContractName() string
	MspId() string
	IdentityX509Certificate() *x509.Certificate
	IdentityEDCSAKey() *ecdsa.PrivateKey
	TlsX509Certificate() *x509.Certificate
	GatewayPeer() string
}

var instance SettingsI
var mtx sync.Mutex = sync.Mutex{}

func SetGlobalSettings(iSettings SettingsI) {
	mtx.Lock()
	instance = iSettings
	mtx.Unlock()
}

func GetGlobalSettings() SettingsI {
	return instance
}

func init() {
	_, ok := os.LookupEnv("BYPASS_ENV_SETTINGS")
	if !ok {
		setting, err := NewSettingsFromEnv()
		if err != nil {
			panic(err)
		}
		SetGlobalSettings(setting)
	}
}
