package utility_sig_graph

import (
	"os"
	"sync"
)

type SettingsI interface {
	PeerAddresses() []string
	ChannelName() string
	ContractName() string
	MspId() string
	X509CertificateData() string
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
