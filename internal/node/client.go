package node

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/easterthebunny/automation-cli/internal/restclient"
	"github.com/easterthebunny/automation-cli/internal/util"
)

const (
	ethKeysEndpoint  = "/v2/keys/eth"
	ocr2KeysEndpoint = "/v2/keys/ocr2"
	p2pKeysEndpoint  = "/v2/keys/p2p"
	csaKeysEndpoint  = "/v2/keys/csa"
)

var (
	ErrAuthentication = fmt.Errorf("authentication failure")
)

// HTTPClient encapsulates all methods used to interact with a chainlink node API.
type HTTPClient interface {
	Get(string, ...map[string]string) (*http.Response, error)
	Post(string, io.Reader) (*http.Response, error)
	Put(string, io.Reader) (*http.Response, error)
	Patch(string, io.Reader, ...map[string]string) (*http.Response, error)
	Delete(string) (*http.Response, error)
}

// authenticate creates a http client with URL, email and password
func authenticate(ctx context.Context, urlStr, email, password string) (*restclient.AuthenticatedHTTPClient, error) {
	remoteNodeURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse url: %s", ErrAuthentication, err.Error())
	}

	opts := restclient.ClientOpts{RemoteNodeURL: *remoteNodeURL}
	request := restclient.SessionRequest{Email: email, Password: password}
	store := &restclient.MemoryCookieStore{}

	tca := restclient.NewSessionCookieAuthenticator(opts, store)
	if _, err = tca.Authenticate(ctx, request); err != nil {
		return nil, fmt.Errorf("%w: session cookie authentication: %s", ErrAuthentication, err.Error())
	}

	return restclient.NewAuthenticatedHTTPClient(opts, tca, request), nil
}

func nodeRequest(client HTTPClient, path string) ([]byte, error) {
	resp, err := client.Get(path)
	if err != nil {
		return []byte{}, fmt.Errorf("GET error from client: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to read response body: %w", err)
	}

	type errorDetail struct {
		Detail string `json:"detail"`
	}

	type errorResp struct {
		Errors []errorDetail `json:"errors"`
	}

	var errs errorResp
	if err := json.Unmarshal(raw, &errs); err == nil && len(errs.Errors) > 0 {
		return []byte{}, fmt.Errorf("error returned from api: %s", errs.Errors[0].Detail)
	}

	return raw, nil
}

type EthKeyPresenter struct {
	Attributes struct {
		Address string `json:"address"`
	} `json:"attributes"`
}

type EthKeyPresenters []EthKeyPresenter

// getNodeAddress returns chainlink node's wallet address
func getNodeAddress(client HTTPClient) (string, error) {
	rawResponse, err := nodeRequest(client, ethKeysEndpoint)
	if err != nil {
		return "", fmt.Errorf("failed to get ETH keys: %w", err)
	}

	var response dataResponse
	if err := json.Unmarshal(rawResponse, &response); err != nil {
		return "", fmt.Errorf("not a data response: %w", err)
	}

	var keys EthKeyPresenters
	if err = json.Unmarshal(response.Data, &keys); err != nil {
		return "", fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return keys[0].Attributes.Address, nil
}

type dataResponse struct {
	Data json.RawMessage `json:"data"`
}

type JAID struct {
	ID string `json:"id"`
}

type P2PKeyPresenter struct {
	JAID
}

type P2PKeyPresenters []P2PKeyPresenter

// getP2PKeyID returns chainlink node's P2P key ID
func getP2PKeyID(client HTTPClient) (string, error) {
	rawResponse, err := nodeRequest(client, p2pKeysEndpoint)
	if err != nil {
		return "", fmt.Errorf("failed to get P2P keys: %w", err)
	}

	var response dataResponse
	if err := json.Unmarshal(rawResponse, &response); err != nil {
		return "", fmt.Errorf("not a data response: %w", err)
	}

	var keys P2PKeyPresenters
	if err = json.Unmarshal(response.Data, &keys); err != nil {
		return "", fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return keys[0].ID, nil
}

type OCR2KeyBundlePresenter struct {
	ID         string `json:"id"`
	Attributes struct {
		ChainType string `json:"chainType"`
	} `json:"attributes"`
}

type OCR2KeyBundlePresenters []OCR2KeyBundlePresenter

// getNodeOCR2Config returns chainlink node's OCR2 bundle key ID
func getNodeOCR2Config(client HTTPClient) (*OCR2KeyBundlePresenter, error) {
	rawResponse, err := nodeRequest(client, ocr2KeysEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get OCR2 keys: %w", err)
	}

	var response dataResponse
	if err := json.Unmarshal(rawResponse, &response); err != nil {
		return nil, fmt.Errorf("not a data response: %w", err)
	}

	var keys OCR2KeyBundlePresenters
	if err = json.Unmarshal(response.Data, &keys); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	var evmKey OCR2KeyBundlePresenter
	for _, key := range keys {
		if key.Attributes.ChainType == "evm" {
			evmKey = key

			break
		}
	}

	return &evmKey, nil
}

// addKeyToKeeper imports the provided ETH sending key to the keeper
func addKeyToKeeper(client HTTPClient, privKeyHex string, chainID int64) (string, error) {
	privkey, err := crypto.HexToECDSA(util.RemoveHexPrefix(privKeyHex))
	if err != nil {
		log.Fatalf("Failed to decode priv key %s: %v", privKeyHex, err)
	}

	address := crypto.PubkeyToAddress(privkey.PublicKey).Hex()

	keyJSON, err := util.FromPrivateKey(privkey).ToEncryptedJSON(DefaultChainlinkNodePassword, util.FastScryptParams)
	if err != nil {
		return "", fmt.Errorf("Failed to encrypt piv key %s: %s", privKeyHex, err.Error())
	}

	importUrl := url.URL{
		Path: "/v2/keys/evm/import",
	}

	query := importUrl.Query()

	query.Set("oldpassword", DefaultChainlinkNodePassword)
	query.Set("evmChainID", fmt.Sprint(chainID))

	importUrl.RawQuery = query.Encode()

	resp, err := client.Post(importUrl.String(), bytes.NewReader(keyJSON))
	if err != nil {
		return "", fmt.Errorf("Failed to import priv key %s: %s", privKeyHex, err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read error response body: %s", err)
		}

		return "", fmt.Errorf("unable import private key: '%v' [%d]", string(body), resp.StatusCode)
	}

	return address, nil
}

type CreateJobRequest struct {
	TOML string `json:"toml"`
}

// createBootstrapJob creates a bootstrap job in the chainlink node by the given address
func createBootstrapJob(client HTTPClient, contractAddr string, chainID int64) error {
	request, err := json.Marshal(CreateJobRequest{
		TOML: fmt.Sprintf(bootstrapJobSpec, contractAddr, chainID),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %s", err)
	}

	resp, err := client.Post("/v2/jobs", bytes.NewReader(request))
	if err != nil {
		return fmt.Errorf("failed to create bootstrap job: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read error response body: %s", err)
		}

		return fmt.Errorf("unable to create bootstrap job: '%v' [%d]", string(body), resp.StatusCode)
	}

	return nil
}

type AutomationJobConfig struct {
	Version           string
	ContractAddr      string
	NodeAddr          string
	BootstrapNodeAddr string
	ChainID           int64
	MercuryCredName   string
}

// createOCR2AutomationJob creates an ocr2keeper job in the chainlink node by the given address
func createOCR2AutomationJob(client HTTPClient, conf AutomationJobConfig) error {
	ocr2KeyConfig, err := getNodeOCR2Config(client)
	if err != nil {
		return fmt.Errorf("failed to get node OCR2 key bundle ID: %s", err)
	}

	request, err := json.Marshal(CreateJobRequest{
		TOML: fmt.Sprintf(ocr2AutomationJobTemplate,
			common.HexToAddress(conf.ContractAddr).Hex(), // contractID
			ocr2KeyConfig.ID,                         // ocrKeyBundleID
			common.HexToAddress(conf.NodeAddr).Hex(), // transmitterID - node wallet address
			conf.BootstrapNodeAddr,                   // bootstrap node key and address
			conf.ChainID,                             // chainID
			conf.Version,                             // contractVersion
			conf.MercuryCredName,                     // mercury credential name
		),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %s", err)
	}

	resp, err := client.Post("/v2/jobs", bytes.NewReader(request))
	if err != nil {
		return fmt.Errorf("failed to create ocr2keeper job: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read error response body: %s", err)
		}

		return fmt.Errorf("unable to create ocr2keeper job: '%s' [%d]", string(body), resp.StatusCode)
	}

	return nil
}
