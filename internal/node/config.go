package node

import (
	"fmt"

	"github.com/easterthebunny/automation-cli/internal/config"
)

const (
	nodeTOML = `[Log]
JSONConsole = true
Level = '%s'
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
LegacyURL = '%s'
URL = '%s'
Username = '%s'
Password = '%s'
`

	ocr2AutomationJobTemplate = `type = "offchainreporting2"
pluginType = "ocr2automation"
relay = "evm"
name = "ocr2-automation"
forwardingAllowed = false
schemaVersion = 1
contractID = "%s"
contractConfigTrackerPollInterval = "15s"
ocrKeyBundleID = "%s"
transmitterID = "%s"
p2pv2Bootstrappers = [
  "%s"
]

[relayConfig]
chainID = %d

[pluginConfig]
maxServiceWorkers = 100
cacheEvictionInterval = "1s"
contractVersion = "%s"
mercuryCredentialName = "%s"`
)

func NodeTOML(conf config.NodeConfig) string {
	return fmt.Sprintf(nodeTOML, conf.LogLevel, conf.ChainID, conf.WSURL, conf.HTTPURL)
}

func SecretTOML(conf config.NodeConfig) string {
	return fmt.Sprintf(secretTOML, conf.MercuryLegacyURL, conf.MercuryURL, conf.MercuryID, conf.MercuryKey)
}
