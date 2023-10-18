package node

import (
	"context"
	"io"
)

type OCR2NodeConfig struct {
	OffChainPublicKey string
	ConfigPublicKey   string
	OnchainPublicKey  string
	P2PKeyID          string
}

func CreateParticipantNode(
	ctx context.Context,
	conf NodeConfig,
	port uint16,
	groupname, name, image string,
	contract, bootstrap, basePath string,
	privateKey *string,
	reset bool,
) (*ChainlinkNode, error) {
	extraTOML := "[P2P]\n[P2P.V2]\nListenAddresses = [\"0.0.0.0:8000\"]"

	var err error

	clNode, err := buildChainlinkNode(
		ctx, io.Discard, conf,
		port, groupname, name, image,
		extraTOML, basePath, reset,
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

func GetParticipantInfo(
	ctx context.Context,
	url string,
) (*OCR2NodeConfig, error) {
	client, err := authenticate(ctx, url, DefaultChainlinkNodeLogin, DefaultChainlinkNodePassword)
	if err != nil {
		return nil, err
	}

	ocr2Conf, err := getNodeOCR2Config(client)
	if err != nil {
		return nil, err
	}

	keyID, err := getP2PKeyID(client)
	if err != nil {
		return nil, err
	}

	return &OCR2NodeConfig{
		OffChainPublicKey: ocr2Conf.Attributes.OffChainPublicKey,
		ConfigPublicKey:   ocr2Conf.Attributes.ConfigPublicKey,
		OnchainPublicKey:  ocr2Conf.Attributes.OnchainPublicKey,
		P2PKeyID:          keyID,
	}, nil
}
