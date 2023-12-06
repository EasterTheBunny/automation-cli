package node

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/easterthebunny/automation-cli/internal/config"
)

const (
	bootstrapJobSpec = `type = "bootstrap"
schemaVersion = 1
name = "ocr2keeper bootstrap node"
contractID = "%s"
relay = "evm"

[relayConfig]
chainID = %d`

	bootstrapTOML = `[P2P]
[P2P.V2]
ListenAddresses = ["0.0.0.0:%s"]`
)

// CreateBootstrapNode starts the ocr2 bootstrap node with the given contract
// address, returns the tcp address of the node.
func CreateBootstrapNode(
	ctx context.Context,
	groupname, registryAddr string,
	conf *config.NodeConfig,
	path string,
	reset bool,
) error {
	node, err := buildChainlinkNode(
		ctx, io.Discard, conf,
		dockerNodeConfig{
			Port:          conf.ListenPort,
			Group:         groupname,
			ContainerName: conf.Name,
			Image:         conf.Image,
			ExtraTOML:     fmt.Sprintf(bootstrapTOML, strconv.Itoa(int(conf.BootstrapListenPort))),
			BasePath:      path,
			Reset:         reset,
		},
	)
	if err != nil {
		return err
	}

	conf.ManagementURL = node.URL()

	client, err := authenticate(ctx, conf.ManagementURL, conf.LoginName, conf.LoginPassword)
	if err != nil {
		return err
	}

	if conf.P2PKeyID, err = getP2PKeyID(client); err != nil {
		return err
	}

	if err = createBootstrapJob(client, registryAddr, conf.ChainID); err != nil {
		return err
	}

	conf.BootstrapAddress = fmt.Sprintf("%s@%s-%s:%d", conf.P2PKeyID, groupname, conf.Name, conf.BootstrapListenPort)

	return nil
}
