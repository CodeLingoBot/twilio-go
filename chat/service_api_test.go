package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/smnalex/twilio-go"
)

var (
	mockClient *mockHTTPClient
	srv        serviceAPI

	ctx   = context.Background()
	setup = func() {
		mockClient = &mockHTTPClient{}
		srv = serviceAPI{mockClient}
	}
)

func TestServiceList(t *testing.T) {

}

func TestServiceRead(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		setup()
		mockClient.GetFunc = func(ctx context.Context, path string) ([]byte, error) {
			if exp := "/Services/SID"; exp != path {
				t.Errorf("exp path %s, got %s", exp, path)
			}

			return ioutil.ReadFile("fixtures/service.json")
		}

		got, err := srv.Read(ctx, "SID")
		if err != nil {
			t.Errorf("exp no err, got %v", err)
		}

		var (
			exp  Service
			f, _ = os.Open("fixtures/service.json")
		)

		if err := json.NewDecoder(f).Decode(&exp); err != nil {
			t.Fatalf("exp no decoding issues")
		}

		if !cmp.Equal(got, exp) {
			t.Errorf("exp a service got diffrent service %v", cmp.Diff(got, exp))
		}
	})

	t.Run("unsuccessful with parsing err", func(t *testing.T) {
		setup()
		mockClient.GetFunc = func(ctx context.Context, path string) ([]byte, error) {
			return ioutil.ReadFile("fixtures/invalid.json")
		}

		if _, got := srv.Read(ctx, "SID"); got == nil {
			t.Errorf("exp parsing err, got %v", got)
		}
	})

	t.Run("unsuccessful with not found err", func(t *testing.T) {
		setup()
		mockClient.GetFunc = func(ctx context.Context, path string) ([]byte, error) {
			return nil, twilio.ErrNotFound
		}

		exp := twilio.ErrNotFound
		if _, got := srv.Read(ctx, "SID"); got != exp {
			t.Errorf("exp err %v, got %v", exp, got)
		}
	})
}

func TestServiceCreate(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		setup()
		mockClient.PostFunc = func(ctx context.Context, path string, body io.Reader) ([]byte, error) {
			if exp := "/Services"; exp != path {
				t.Errorf("exp path %s, got %s", exp, path)
			}

			var (
				gotBody, _ = ioutil.ReadAll(body)
				expBody    = []byte(`{"friendly_name":"hello there"}`)
			)

			if bytes.Compare(gotBody, expBody) != 0 {
				t.Errorf("exp body %s, got %s", expBody, gotBody)
			}

			return ioutil.ReadFile("fixtures/service.json")
		}

		updateParams := ServiceUpdateParams{FriendlyName: "hello there"}
		got, err := srv.Create(ctx, updateParams)
		if err != nil {
			t.Errorf("exp no err, got %v", err)
		}

		var (
			exp  Service
			f, _ = os.Open("fixtures/service.json")
		)

		if err := json.NewDecoder(f).Decode(&exp); err != nil {
			t.Fatalf("exp no decoding issues")
		}

		if !cmp.Equal(got, exp) {
			t.Errorf("exp a service got diffrent service %v", cmp.Diff(got, exp))
		}
	})

	t.Run("unsuccessful with parsing err", func(t *testing.T) {
		setup()
		mockClient.PostFunc = func(ctx context.Context, path string, body io.Reader) ([]byte, error) {
			return ioutil.ReadFile("fixtures/invalid.json")
		}

		updateParams := ServiceUpdateParams{FriendlyName: "hello there"}
		_, got := srv.Create(ctx, updateParams)
		if got == nil {
			t.Errorf("exp parsing err, got %v", got)
		}
	})

	t.Run("successful with request err", func(t *testing.T) {
		setup()

		mockClient.PostFunc = func(ctx context.Context, path string, body io.Reader) ([]byte, error) {
			return nil, errors.New("test")
		}

		updateParams := ServiceUpdateParams{FriendlyName: "hello there"}
		_, got := srv.Create(ctx, updateParams)
		if got == nil {
			t.Errorf("exp err(any), got %v", got)
		}
	})
}

func TestServiceUpdate(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		setup()
		mockClient.PostFunc = func(ctx context.Context, path string, body io.Reader) ([]byte, error) {
			if exp := "/Services/SID"; exp != path {
				t.Errorf("exp path %s, got %s", exp, path)
			}
			return ioutil.ReadFile("fixtures/service.json")
		}

		updateParams := ServiceUpdateParams{FriendlyName: "hello there"}
		_, got := srv.Update(ctx, "SID", updateParams)
		if got != nil {
			t.Errorf("exp no err, got %v", got)
		}
	})
}

func TestServiceDelete(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		setup()
		mockClient.DeleteFunc = func(ctx context.Context, path string) ([]byte, error) {
			if exp := "/Services/SID"; exp != path {
				t.Errorf("exp path %s, got %s", exp, path)
			}
			return nil, nil
		}

		err := srv.Delete(ctx, "SID")
		if err != nil {
			t.Errorf("exp no err, got %v", err)
		}
	})

	t.Run("unsuccessful with request err", func(t *testing.T) {
		setup()
		mockClient.DeleteFunc = func(ctx context.Context, path string) ([]byte, error) {
			return nil, errors.New("")
		}

		got := srv.Delete(ctx, "SID")
		if got == nil {
			t.Errorf("exp err(any), got %v", got)
		}
	})
}

type mockHTTPClient struct {
	GetFunc    func(context.Context, string) ([]byte, error)
	PostFunc   func(context.Context, string, io.Reader) ([]byte, error)
	DeleteFunc func(context.Context, string) ([]byte, error)
}

func (m *mockHTTPClient) Get(ctx context.Context, path string) ([]byte, error) {
	return m.GetFunc(ctx, path)
}

func (m *mockHTTPClient) Post(ctx context.Context, path string, body io.Reader) ([]byte, error) {
	return m.PostFunc(ctx, path, body)
}

func (m *mockHTTPClient) Delete(ctx context.Context, path string) ([]byte, error) {
	return m.DeleteFunc(ctx, path)
}
