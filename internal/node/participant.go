package node

import (
	"context"
	"fmt"
	"io"

	"github.com/easterthebunny/automation-cli/internal/config"
	"github.com/easterthebunny/automation-cli/internal/restclient"
)

type OCR2NodeConfig struct {
	OffChainPublicKey string
	ConfigPublicKey   string
	OnchainPublicKey  string
	P2PKeyID          string
}

func CreateParticipantNode(
	ctx context.Context,
	groupname, registryAddr string,
	bootstrap config.NodeConfig,
	conf *config.NodeConfig,
	basePath string,
	privateKey *string,
	reset bool,
) error {
	extraTOML := fmt.Sprintf("[P2P]\n[P2P.V2]\nListenAddresses = [\"0.0.0.0:%d\"]", bootstrap.BootstrapListenPort)

	var err error

	clNode, err := buildChainlinkNode(
		ctx, io.Discard, conf,
		dockerNodeConfig{
			Port:          conf.ListenPort,
			Group:         groupname,
			ContainerName: conf.Name,
			Image:         conf.Image,
			ExtraTOML:     extraTOML,
			BasePath:      basePath,
			Reset:         reset,
		},
	)
	if err != nil {
		return err
	}

	conf.ManagementURL = clNode.URL()

	client, err := authenticate(ctx, conf.ManagementURL, conf.LoginName, conf.LoginPassword)
	if err != nil {
		return err
	}

	// get or set address
	if privateKey != nil {
		addr, err := addKeyToKeeper(client, *privateKey, conf.LoginPassword, conf.ChainID)
		if err != nil {
			return err
		}

		clNode.Address = addr
	} else {
		addr, err := getNodeAddress(client)
		if err != nil {
			return err
		}

		clNode.Address = addr
	}

	conf.Address = clNode.Address

	// create automation job
	if err := createOCR2AutomationJob(client, AutomationJobConfig{
		Version:           "v2.1",
		ContractAddr:      registryAddr,
		NodeAddr:          clNode.Address,
		BootstrapNodeAddr: bootstrap.BootstrapAddress,
		ChainID:           conf.ChainID,
		MercuryCredName:   "cred1",
	}); err != nil {
		return err
	}

	return getParticipantInfo(client, conf)
}

func getParticipantInfo(
	client *restclient.AuthenticatedHTTPClient,
	conf *config.NodeConfig,
) error {
	ocr2Conf, err := getNodeOCR2Config(client)
	if err != nil {
		return err
	}

	keyID, err := getP2PKeyID(client)
	if err != nil {
		return err
	}

	conf.OffChainPublicKey = ocr2Conf.Attributes.OffChainPublicKey
	conf.ConfigPublicKey = ocr2Conf.Attributes.ConfigPublicKey
	conf.OnchainPublicKey = ocr2Conf.Attributes.OnchainPublicKey
	conf.P2PKeyID = keyID

	return nil
}

func RemoveParticipantNode(
	ctx context.Context,
	groupname string,
	conf config.NodeConfig,
) error {
	return removeChainlinkNode(
		ctx,
		dockerNodeConfig{
			Port:          conf.ListenPort,
			Group:         groupname,
			ContainerName: conf.Name,
			Image:         conf.Image,
		},
	)
}
