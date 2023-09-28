package node

import (
	"context"
	"fmt"
	"io"
	"strconv"
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
	conf NodeConfig,
	groupname, image, addr string,
	uiPort, p2pv2Port int,
	path string,
	reset bool,
) (string, error) {
	const containerName = "bootstrap"

	node, err := buildChainlinkNode(
		ctx, io.Discard, conf,
		uint16(uiPort), groupname, containerName, image,
		fmt.Sprintf(bootstrapTOML, strconv.Itoa(p2pv2Port)), path, reset,
	)
	if err != nil {
		return "", err
	}

	urlRaw := node.URL()

	client, err := authenticate(ctx, urlRaw, DefaultChainlinkNodeLogin, DefaultChainlinkNodePassword)
	if err != nil {
		return "", err
	}

	p2pKeyID, err := getP2PKeyID(client)
	if err != nil {
		return "", err
	}

	if err = createBootstrapJob(client, addr, conf.ChainID); err != nil {
		return "", err
	}

	tcpAddr := fmt.Sprintf("%s@%s:%d", p2pKeyID, containerName, p2pv2Port)

	return tcpAddr, nil
}
