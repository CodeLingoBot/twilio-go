package twilio

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

var (
	// ErrNotFound resource not found
	ErrNotFound = errors.New("not found")
)

// HTTPClient http wrapper which allows interacting with various twilio apis.
type HTTPClient interface {
	Get(context.Context, string) ([]byte, error)
	Post(context.Context, string, io.Reader) ([]byte, error)
	Delete(context.Context, string) ([]byte, error)
}

// RequestHandler abstracts `http.Client`
type RequestHandler interface {
	Do(*http.Request) (*http.Response, error)
}

type apiClient struct {
	url        *url.URL
	accountSID string
	authToken  string
	RequestHandler
}

// NewHTTPClient returns a new twilio.Client which can be used to access various API
// twilio rest apis. It requires a custom type `twilio.RequestHandler` which has the
// method signature of the `http.Client` struct `Do` method.
func NewHTTPClient(accountSID, authToken, baseURL string, rh RequestHandler) (HTTPClient, error) {
	url, err := url.Parse(baseURL)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse url")
	}

	return &apiClient{
		url:            url,
		accountSID:     accountSID,
		authToken:      authToken,
		RequestHandler: rh,
	}, nil
}

func (client *apiClient) Get(ctx context.Context, path string) ([]byte, error) {
	return client.request(ctx, http.MethodGet, path, nil)
}

func (client *apiClient) Post(ctx context.Context, path string, body io.Reader) ([]byte, error) {
	return client.request(ctx, http.MethodPost, path, body)
}

func (client *apiClient) Delete(ctx context.Context, path string) ([]byte, error) {
	return client.request(ctx, http.MethodDelete, path, nil)
}

func (client *apiClient) request(ctx context.Context, method, path string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, client.url.String()+path, body)
	if err != nil {
		return nil, errors.Wrap(err, "could not create request")
	}

	{
		req.SetBasicAuth(client.accountSID, client.authToken)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.WithContext(ctx)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get a response for %s", req.URL)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode < 200 || 400 <= statusCode {
		if statusCode == http.StatusNotFound {
			return nil, ErrNotFound
		}

		return nil, errors.Errorf("unexpected status code: %d", statusCode)
	}

	return ioutil.ReadAll(resp.Body)
}
