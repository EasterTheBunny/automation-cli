package node

import (
	"context"
	"io"
)

func CreatParticipantNode(
	ctx context.Context,
	conf NodeConfig,
	port uint16,
	groupname, name, image, addr string,
) (string, error) {
	extraTOML := "[P2P]\n[P2P.V2]\nListenAddresses = [\"0.0.0.0:8000\"]"

	// Run chainlink nodes and create jobs
	// Run chainlink node
	var err error
	clNode, err := buildChainlinkNode(
		ctx, io.Discard, conf,
		port, groupname, name, image,
		extraTOML, false,
	)
	if err != nil {
		return "", err
	}

	return clNode.URL(), nil
}
