package domain

const (
	Registrar                 = "registrar"
	Registry                  = "registry"
	VerifiableLoadLogTrigger  = "verifiable-load-log-trigger"
	VerifiableLoadConditional = "verifiable-load-conditional"
	LinkToken                 = "link-token"
	LinkEthFeed               = "link-eth-feed"
)

var (
	ContractNames = []string{
		Registrar,
		Registry,
		VerifiableLoadLogTrigger,
		VerifiableLoadConditional,
		LinkToken,
		LinkEthFeed,
	}
)
