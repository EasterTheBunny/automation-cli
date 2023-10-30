package restclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/smartcontractkit/chainlink/v2/core/store/models"
)

var (
	ErrEncoding       = fmt.Errorf("encoding")
	ErrAuthentication = fmt.Errorf("authentication")
	ErrConnection     = fmt.Errorf("connection")
)

type SessionRequest struct {
	Email    string `json:"Email"`
	Password string `json:"Password"`
}

type ClientOpts struct {
	RemoteNodeURL      url.URL
	InsecureSkipVerify bool
}

// SessionCookieAuthenticator is a concrete implementation of CookieAuthenticator
// that retrieves a session id for the user with credentials from the session request.
type SessionCookieAuthenticator struct {
	config ClientOpts
	store  CookieStore
	logger *log.Logger
}

// NewSessionCookieAuthenticator creates a SessionCookieAuthenticator using the passed config
// and builder.
func NewSessionCookieAuthenticator(
	config ClientOpts,
	store CookieStore,
) *SessionCookieAuthenticator {
	return &SessionCookieAuthenticator{config: config, store: store}
}

// Cookie Returns the previously saved authentication cookie.
func (t *SessionCookieAuthenticator) Cookie() (*http.Cookie, error) {
	cookie, err := t.store.Retrieve()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to retrieve stored cookie: %s", ErrAuthentication, err.Error())
	}

	return cookie, nil
}

// Authenticate retrieves a session ID via a cookie and saves it to disk.
func (t *SessionCookieAuthenticator) Authenticate(ctx context.Context, request SessionRequest) (*http.Cookie, error) {
	bRequest := new(bytes.Buffer)

	err := json.NewEncoder(bRequest).Encode(request)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to encode auth request: %s", ErrEncoding, err.Error())
	}

	url := t.config.RemoteNodeURL.String() + "/sessions"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bRequest)
	if err != nil {
		return nil, fmt.Errorf("%w: http request initialization failed: %s", ErrConnection, err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	client := newHTTPClient(t.config.InsecureSkipVerify)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: http authentication request failed: %s", ErrConnection, err.Error())
	}

	defer resp.Body.Close()

	_, err = parseResponse(resp)
	if err != nil {
		return nil, err
	}

	cookies := resp.Cookies()
	if len(cookies) == 0 {
		return nil, fmt.Errorf("%w: did not receive cookie with session id", ErrAuthentication)
	}

	sc := findSessionCookie(cookies)
	if err := t.store.Save(sc); err != nil {
		return nil, fmt.Errorf("%w: failed to store cookie: %s", ErrConnection, err.Error())
	}

	return sc, nil
}

// Deletes any stored session.
func (t *SessionCookieAuthenticator) Logout() error {
	if err := t.store.Reset(); err != nil {
		return fmt.Errorf("%w: failed to reset cookie store: %s", ErrConnection, err.Error())
	}

	return nil
}

// CookieStore is a place to store and retrieve cookies.
type CookieStore interface {
	Save(cookie *http.Cookie) error
	Retrieve() (*http.Cookie, error)
	Reset() error
}

// MemoryCookieStore keeps a single cookie in memory.
type MemoryCookieStore struct {
	Cookie *http.Cookie
}

// Save stores a cookie.
func (m *MemoryCookieStore) Save(cookie *http.Cookie) error {
	m.Cookie = cookie

	return nil
}

// Removes any stored cookie.
func (m *MemoryCookieStore) Reset() error {
	m.Cookie = nil

	return nil
}

// Retrieve returns any Saved cookies.
func (m *MemoryCookieStore) Retrieve() (*http.Cookie, error) {
	return m.Cookie, nil
}

// parseErrorResponseBody parses response body from web API and returns a single string containing all errors.
func parseErrorResponseBody(responseBody []byte) (string, error) {
	if responseBody == nil {
		return "Empty error message", nil
	}

	var errors models.JSONAPIErrors

	if err := json.Unmarshal(responseBody, &errors); err != nil || len(errors.Errors) == 0 {
		return "", fmt.Errorf("%w: failed to unmarshal response: %s", ErrEncoding, err.Error())
	}

	var errorDetails strings.Builder

	errorDetails.WriteString(errors.Errors[0].Detail)

	for _, errorDetail := range errors.Errors[1:] {
		fmt.Fprintf(&errorDetails, "\n%s", errorDetail.Detail)
	}

	return errorDetails.String(), nil
}

func parseResponse(resp *http.Response) ([]byte, error) {
	bBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return bBody, fmt.Errorf("%w: connection failure with status (%s): %s", ErrConnection, resp.Status, err.Error())
	}
	fmt.Println(string(bBody))

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return bBody, fmt.Errorf("%w: status unauthorized", ErrAuthentication)
	case http.StatusForbidden:
		return bBody, fmt.Errorf("%w: forbidden", ErrAuthentication)
	case http.StatusBadRequest:
		errorMessage, err := parseErrorResponseBody(bBody)
		if err != nil {
			return bBody, err
		}

		return bBody, fmt.Errorf("%w: bad request with message: %s", ErrAuthentication, errorMessage)
	default:
		return bBody, nil
	}
}

func findSessionCookie(cookies []*http.Cookie) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == "clsession" {
			return cookie
		}
	}

	return nil
}
