package twilio

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
)

var (
	mockedRequestHandler *mockRequestHandler
	client               HTTPClient

	baseURL = "https://twilio.com/v2"
	ctx     = context.Background()
	auth    = "auth"
	acc     = "acc"
	setup   = func() {
		mockedRequestHandler = &mockRequestHandler{}
		client, _ = NewClient(acc, auth, baseURL, mockedRequestHandler)
	}
)

func TestNewClient(t *testing.T) {
	t.Run("correct configuration", func(t *testing.T) {
		if _, err := NewClient(acc, auth, baseURL, nil); err != nil {
			t.Errorf("exp no err, got %v", err)
		}
	})

	t.Run("invalid base URL", func(t *testing.T) {
		if _, err := NewClient(acc, auth, "%", nil); err == nil {
			t.Errorf("exp invalid URL parsing err, got %v", err)
		}
	})
}

func TestGet(t *testing.T) {
	var path = "/get"

	t.Run("successful request", func(t *testing.T) {
		setup()

		type ctxType string
		var ctxKey = ctxType("key")

		mockedRequestHandler.requestHandlerFunc = func(r *http.Request) (*http.Response, error) {
			if exp, got := baseURL+path, r.URL.String(); exp != got {
				t.Errorf("exp path %s, got %s", exp, r.URL.String())
			}
			if usr, pass, ok := r.BasicAuth(); !ok || usr != acc || pass != auth {
				t.Errorf("exp req basic auth with %s:%s, got %s:%s", acc, auth, usr, pass)
			}
			if got := ctx.Value(ctxKey); got == nil {
				t.Error("exp ctx val, got none")
			}
			exp := "application/x-www-form-urlencoded"
			if got := r.Header.Get("Content-Type"); exp != got {
				t.Errorf("exp header %s, got %s", exp, got)
			}
			body := ioutil.NopCloser(strings.NewReader("{}"))
			return &http.Response{StatusCode: 200, Body: body}, nil
		}

		ctx = context.WithValue(ctx, ctxKey, "test")
		resp, err := client.Get(ctx, path)
		if err != nil {
			t.Errorf("exp no err, got %v", err)
		}
		if exp := []byte("{}"); !cmp.Equal(resp, exp) {
			t.Errorf("exp resp %s, got %s", exp, resp)
		}
		if !mockedRequestHandler.requestInvoked {
			t.Error("exp HTTPClient.Get to be invoked")
		}
	})

	t.Run("unsuccessful request with err", func(t *testing.T) {
		setup()
		var respErr = errors.New("test")
		mockedRequestHandler.requestHandlerFunc = func(r *http.Request) (*http.Response, error) {
			return &http.Response{}, respErr
		}

		_, err := client.Get(ctx, path)
		if err == nil {
			t.Fatal("exp err, got none")
		}
		if got := errors.Cause(err); respErr != got {
			t.Errorf("exp err %v, got %v", respErr, got)
		}
		if !mockedRequestHandler.requestInvoked {
			t.Error("exp HTTPClient.Get to be invoked")
		}
	})

	t.Run("unsuccessful request status 404", func(t *testing.T) {
		setup()
		mockedRequestHandler.requestHandlerFunc = func(r *http.Request) (*http.Response, error) {
			body := ioutil.NopCloser(strings.NewReader("{}"))
			return &http.Response{StatusCode: 404, Body: body}, nil
		}

		if _, err := client.Get(ctx, path); err != ErrNotFound {
			t.Errorf("exp err not found, got %v", err)
		}
		if !mockedRequestHandler.requestInvoked {
			t.Error("exp HTTPClient.Get to be invoked")
		}
	})

	t.Run("unsuccessful request status 5xx", func(t *testing.T) {
		setup()
		mockedRequestHandler.requestHandlerFunc = func(r *http.Request) (*http.Response, error) {
			body := ioutil.NopCloser(strings.NewReader("{}"))
			return &http.Response{StatusCode: 500, Body: body}, nil
		}

		exp := fmt.Sprintf("unexpected status code: %d", 500)
		if _, err := client.Get(ctx, path); exp != err.Error() {
			t.Errorf("exp err not found, got %v", err)
		}
		if !mockedRequestHandler.requestInvoked {
			t.Error("exp HTTPClient.Get to be invoked")
		}
	})

	t.Run("unsuccessful with invalid req url err", func(t *testing.T) {
		setup()
		path = "/get%2"
		mockedRequestHandler.requestHandlerFunc = func(r *http.Request) (*http.Response, error) {
			body := ioutil.NopCloser(strings.NewReader("{}"))
			return &http.Response{StatusCode: 500, Body: body}, nil
		}

		if _, err := client.Get(ctx, path); err == nil {
			t.Errorf("exp parsing err, got %v", err)
		}
		if mockedRequestHandler.requestInvoked {
			t.Error("exp HTTPClient.Get to be invoked")
		}
	})
}

func TestPost(t *testing.T) {
	setup()

	var (
		path = "/post"
		body = []byte("{}")
	)

	mockedRequestHandler.requestHandlerFunc = func(r *http.Request) (*http.Response, error) {
		if exp, got := baseURL+path, r.URL.String(); exp != got {
			t.Errorf("exp path %s, got %s", exp, got)
		}
		reqBody, _ := ioutil.ReadAll(r.Body)
		if exp, got := body, reqBody; !cmp.Equal(got, exp) {
			t.Errorf("exp resp %s, got %s", exp, got)
		}
		body := ioutil.NopCloser(bytes.NewReader(body))
		return &http.Response{StatusCode: 200, Body: body}, nil
	}

	resp, err := client.Post(ctx, path, bytes.NewReader(body))
	if err != nil {
		t.Errorf("exp not err, got %v", err)
	}
	if exp := body; !cmp.Equal(resp, exp) {
		t.Errorf("exp resp %s, got %s", exp, resp)
	}
}

func TestDelete(t *testing.T) {
	setup()

	var path = "/delete"
	mockedRequestHandler.requestHandlerFunc = func(r *http.Request) (*http.Response, error) {
		if exp, got := baseURL+path, r.URL.String(); exp != got {
			t.Errorf("exp path %s, got %s", exp, got)
		}
		body := ioutil.NopCloser(strings.NewReader("{}"))
		return &http.Response{StatusCode: 200, Body: body}, nil
	}

	resp, err := client.Delete(ctx, path)
	if err != nil {
		t.Errorf("exp not err, got %v", err)
	}
	if exp := []byte("{}"); !cmp.Equal(resp, exp) {
		t.Errorf("exp resp %s, got %s", exp, resp)
	}
}

type mockRequestHandler struct {
	requestInvoked     bool
	requestHandlerFunc func(*http.Request) (*http.Response, error)
}

func (m *mockRequestHandler) Do(r *http.Request) (*http.Response, error) {
	m.requestInvoked = true
	return m.requestHandlerFunc(r)
}
