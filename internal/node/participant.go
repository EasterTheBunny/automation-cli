package node

import (
	"context"
	"io"
)

func CreatParticipantNode(
	ctx context.Context,
	conf NodeConfig,
	port uint16,
	groupname, name, image string,
	contract, bootstrap, basePath string,
	privateKey *string,
) (*ChainlinkNode, error) {
	extraTOML := "[P2P]\n[P2P.V2]\nListenAddresses = [\"0.0.0.0:8000\"]"

	var err error

	clNode, err := buildChainlinkNode(
		ctx, io.Discard, conf,
		port, groupname, name, image,
		extraTOML, basePath, false,
	)
	if err != nil {
		return nil, err
	}

	urlRaw := clNode.URL()

	client, err := authenticate(ctx, urlRaw, DefaultChainlinkNodeLogin, DefaultChainlinkNodePassword)
	if err != nil {
		return nil, err
	}

	// get or set address
	if privateKey != nil {
		addr, err := addKeyToKeeper(client, *privateKey, conf.ChainID)
		if err != nil {
			return nil, err
		}

		clNode.Address = addr
	} else {
		addr, err := getNodeAddress(client)
		if err != nil {
			return nil, err
		}

		clNode.Address = addr
	}

	// create automation job
	if err := createOCR2AutomationJob(client, AutomationJobConfig{
		Version:           "v2.1",
		ContractAddr:      contract,
		NodeAddr:          clNode.Address,
		BootstrapNodeAddr: bootstrap,
		ChainID:           conf.ChainID,
		MercuryCredName:   "cred1",
	}); err != nil {
		return nil, err
	}

	return clNode, nil
}
