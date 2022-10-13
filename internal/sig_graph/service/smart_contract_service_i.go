package service_sig_graph

type SmartContractServiceI interface {
	CreateTransaction(
		iFunctionName string,
		iArgs ...string,
	) (string, error)

	Query(
		iFunctionName string,
		iArgs ...string,
	) (string, error)
}
