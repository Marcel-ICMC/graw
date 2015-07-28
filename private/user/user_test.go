package user

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/paytonturnage/graw/data"
	"github.com/paytonturnage/graw/private/auth"
	"github.com/paytonturnage/graw/private/client"
	"github.com/paytonturnage/graw/private/testutil"
)

func TestNew(t *testing.T) {
	agent := &data.UserAgent{}
	if err := proto.UnmarshalText(`
		user_agent: "agent"
		client_id: "id"
		client_secret: "secret"
		username: "username"
		password: "password"
	`, agent); err != nil {
		t.Fatalf("failed to build expectation proto: %v", err)
	}
	expected := &User{
		agent: "agent",
		authorizer: auth.NewOAuth2Authorizer(
			"id",
			"secret",
			"username",
			"password"),
		client: nil,
	}
	actual := New(agent)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(
			"user built incorrectly; got %v, wanted %v",
			actual,
			expected)
	}

}

func TestAuth(t *testing.T) {
	user := &User{}
	user.authorizer = auth.NewMockAuth(nil, fmt.Errorf("BAD THING"))
	if err := user.Auth(); err == nil {
		t.Error("auth did not return an error on failure")
	}

	user.authorizer = auth.NewMockAuth(http.DefaultClient, nil)
	if err := user.Auth(); err != nil {
		t.Error("auth returned error on success")
	}
	if user.client != http.DefaultClient {
		t.Error("client not set upon successful auth")
	}
}

func TestExec(t *testing.T) {
	user := &User{}
	req := &http.Request{}
	expected := &struct {
		Key string
	}{Key: "value"}
	actual := &struct {
		Key string `"json:"key,omitempty"`
	}{}
	jsonAgent := `{"key": "value"}`

	user.client = client.NewMockClient(
		&http.Response{
			StatusCode: 200,
			Body:       testutil.NewReadCloser(jsonAgent, nil),
		},
		fmt.Errorf("A BAD THING HAPPENED"),
	)
	if err := user.Exec(req, actual); err == nil {
		t.Error("failed request did not return an error")
	}

	user.client = client.NewMockClient(
		&http.Response{
			StatusCode: 200,
			Body: testutil.NewReadCloser(
				jsonAgent,
				fmt.Errorf("misbehavior bad stuff")),
		},
		nil,
	)
	if err := user.Exec(req, actual); err == nil {
		t.Error("corrupt actualonse body did not return an error")
	}

	user.client = client.NewMockClient(
		&http.Response{
			StatusCode: 201,
			Body:       testutil.NewReadCloser(jsonAgent, nil),
		},
		nil,
	)
	if err := user.Exec(req, actual); err == nil {
		t.Error("bad status code did not return an error")
	}

	user.client = client.NewMockClient(
		&http.Response{
			StatusCode: 200,
			Body:       nil,
		},
		nil,
	)
	if err := user.Exec(req, actual); err == nil {
		t.Error("nil body did not return an error")
	}

	user.client = client.NewMockClient(
		&http.Response{
			StatusCode: 200,
			Body:       testutil.NewReadCloser(jsonAgent, nil),
		},
		nil,
	)
	err := user.Exec(req, actual)
	if err != nil {
		t.Errorf("exec failed: %v", err)
	}
	if actual.Key != expected.Key {
		t.Errorf(
			"response incorrect; got %v, wanted %v",
			actual,
			expected)
	}
}
