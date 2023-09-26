package node

import "fmt"

const (
	DefaultChainlinkNodePassword = "fj293fbBnlQ!f9vNs~#"
	DefaultChainlinkNodeLogin    = "notreal@fakeemail.ch"
)

const (
	nodeTOML = `[Log]
JSONConsole = true
Level = 'debug'
[WebServer]
AllowOrigins = '*'
SecureCookies = false
SessionTimeout = '999h0m0s'
[WebServer.TLS]
HTTPSPort = 0
[Feature]
LogPoller = true
[OCR2]
Enabled = true
[P2P]
[P2P.V2]
Enabled = true
[Keeper]
TurnLookBack = 0
[[EVM]]
ChainID = '%d'
[[EVM.Nodes]]
Name = 'node-0'
WSURL = '%s'
HTTPURL = '%s'
`
	secretTOML = `
[Mercury.Credentials.cred1]
URL = '%s'
Username = '%s'
Password = '%s'
`
)

type NodeConfig struct {
	ChainID     int64
	NodeWSSURL  string
	NodeHttpURL string

	MercuryURL string
	MercuryID  string
	MercuryKey string
}

func NodeTOML(conf NodeConfig) string {
	return fmt.Sprintf(nodeTOML, conf.ChainID, conf.NodeWSSURL, conf.NodeHttpURL)
}

func SecretTOML(conf NodeConfig) string {
	return fmt.Sprintf(secretTOML, conf.MercuryURL, conf.MercuryID, conf.MercuryKey)
}
