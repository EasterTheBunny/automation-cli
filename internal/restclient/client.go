package restclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// CookieAuthenticator is the interface to generating a cookie to authenticate
// future HTTP requests.
type CookieAuthenticator interface {
	Cookie() (*http.Cookie, error)
	Authenticate(context.Context, SessionRequest) (*http.Cookie, error)
	Logout() error
}

type AuthenticatedHTTPClient struct {
	client         *http.Client
	cookieAuth     CookieAuthenticator
	sessionRequest SessionRequest
	remoteNodeURL  url.URL
}

// NewAuthenticatedHTTPClient uses the CookieAuthenticator to generate a sessionID
// which is then used for all subsequent HTTP API requests.
func NewAuthenticatedHTTPClient(
	clientOpts ClientOpts,
	cookieAuth CookieAuthenticator,
	sessionRequest SessionRequest,
) *AuthenticatedHTTPClient {
	return &AuthenticatedHTTPClient{
		client:         newHTTPClient(clientOpts.InsecureSkipVerify),
		cookieAuth:     cookieAuth,
		sessionRequest: sessionRequest,
		remoteNodeURL:  clientOpts.RemoteNodeURL,
	}
}

func newHTTPClient(insecureSkipVerify bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify}, //nolint:gosec
	}

	return &http.Client{Transport: tr}
}

// Get performs an HTTP Get using the authenticated HTTP client's cookie.
func (h *AuthenticatedHTTPClient) Get(path string, headers ...map[string]string) (*http.Response, error) {
	return h.doRequest("GET", path, nil, headers...)
}

// Post performs an HTTP Post using the authenticated HTTP client's cookie.
func (h *AuthenticatedHTTPClient) Post(path string, body io.Reader) (*http.Response, error) {
	return h.doRequest("POST", path, body)
}

// Put performs an HTTP Put using the authenticated HTTP client's cookie.
func (h *AuthenticatedHTTPClient) Put(path string, body io.Reader) (*http.Response, error) {
	return h.doRequest("PUT", path, body)
}

// Patch performs an HTTP Patch using the authenticated HTTP client's cookie.
func (h *AuthenticatedHTTPClient) Patch(
	path string,
	body io.Reader,
	headers ...map[string]string,
) (*http.Response, error) {
	return h.doRequest("PATCH", path, body, headers...)
}

// Delete performs an HTTP Delete using the authenticated HTTP client's cookie.
func (h *AuthenticatedHTTPClient) Delete(path string) (*http.Response, error) {
	return h.doRequest("DELETE", path, nil)
}

//nolint:cyclop
func (h *AuthenticatedHTTPClient) doRequest(
	verb, path string,
	body io.Reader,
	headerArgs ...map[string]string,
) (*http.Response, error) {
	var headers map[string]string
	if len(headerArgs) > 0 {
		headers = headerArgs[0]
	} else {
		headers = map[string]string{}
	}

	request, err := http.NewRequestWithContext(context.Background(), verb, h.remoteNodeURL.String()+path, body)
	if err != nil {
		return nil, fmt.Errorf("%w: http request initialization failed: %s", ErrConnection, err.Error())
	}

	request.Header.Set("Content-Type", "application/json")

	for key, value := range headers {
		request.Header.Add(key, value)
	}

	cookie, err := h.cookieAuth.Cookie()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get authentication cookie: %s", ErrConnection, err.Error())
	} else if cookie != nil {
		request.AddCookie(cookie)
	}

	response, err := h.client.Do(request)
	if err != nil {
		return response, fmt.Errorf("%w: http request failed (%s): %s", ErrConnection, path, err.Error())
	}

	if response.StatusCode == http.StatusUnauthorized &&
		(h.sessionRequest.Email != "" || h.sessionRequest.Password != "") {
		var cookieerr error

		cookie, cookieerr = h.cookieAuth.Authenticate(context.Background(), h.sessionRequest)
		if cookieerr != nil {
			return response, fmt.Errorf("%w: cookie authentication failed: %s", ErrAuthentication, cookieerr.Error())
		}

		request.Header.Set("Cookie", "")
		request.AddCookie(cookie)

		response, err = h.client.Do(request)
		if err != nil {
			return response, fmt.Errorf("%w: http request failed (%s): %s", ErrConnection, path, err.Error())
		}
	}

	return response, nil
}
