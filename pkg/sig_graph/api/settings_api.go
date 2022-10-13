package api_sig_graph

import utility_sig_graph "sig_graph_scp/internal/sig_graph/utility"

type SettingsI interface {
	utility_sig_graph.SettingsI
}

func SetGlobalSettings(iSettings SettingsI) {
	utility_sig_graph.SetGlobalSettings(iSettings)
}

func GetGlobalSettings() SettingsI {
	return utility_sig_graph.GetGlobalSettings()
}
