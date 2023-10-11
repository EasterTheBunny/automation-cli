package domain

const (
	Registrar                 = "registrar"
	Registry                  = "registry"
	VerifiableLoadLogTrigger  = "verifiable-load-log-trigger"
	VerifiableLoadConditional = "verifiable-load-conditional"
)

var (
	ContractNames = []string{
		Registrar,
		Registry,
		VerifiableLoadLogTrigger,
		VerifiableLoadConditional,
	}
)
