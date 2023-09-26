package node

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/easterthebunny/automation-cli/internal/restclient"
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
