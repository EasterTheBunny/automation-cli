package domain

const (
	Registrar                 = "registrar"
	Registry                  = "registry"
	VerifiableLoadLogTrigger  = "log-trigger"
	VerifiableLoadConditional = "conditional"
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
