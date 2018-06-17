package apigo

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/tj/assert"
)

func TestNewRequest_path(t *testing.T) {
	e := events.APIGatewayProxyRequest{
		Path: "/pets/luna",
	}

	r, err := DefaultProxy(context.Background(), e)
	assert.NoError(t, err)

	assert.Equal(t, "GET", r.Method)
	assert.Equal(t, `/pets/luna`, r.URL.Path)
	assert.Equal(t, `/pets/luna`, r.URL.String())
}

func TestNewRequest_method(t *testing.T) {
	e := events.APIGatewayProxyRequest{
		HTTPMethod: "DELETE",
		Path:       "/pets/luna",
	}

	r, err := DefaultProxy(context.Background(), e)
	assert.NoError(t, err)

	assert.Equal(t, "DELETE", r.Method)
}

func TestNewRequest_queryString(t *testing.T) {
	e := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/pets",
		QueryStringParameters: map[string]string{
			"order":  "desc",
			"fields": "name,species",
		},
	}

	r, err := DefaultProxy(context.Background(), e)
	assert.NoError(t, err)

	assert.Equal(t, `/pets?fields=name%2Cspecies&order=desc`, r.URL.String())
	assert.Equal(t, `desc`, r.URL.Query().Get("order"))
}

func TestNewRequest_remoteAddr(t *testing.T) {
	e := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/pets",
		RequestContext: events.APIGatewayProxyRequestContext{
			Identity: events.APIGatewayRequestIdentity{
				SourceIP: "1.2.3.4",
			},
		},
	}

	r, err := DefaultProxy(context.Background(), e)
	assert.NoError(t, err)

	assert.Equal(t, `1.2.3.4`, r.RemoteAddr)
}

func TestNewRequest_header(t *testing.T) {
	e := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Path:       "/pets",
		Body:       `{ "name": "Tobi" }`,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"X-Foo":        "bar",
			"Host":         "example.com",
		},
		RequestContext: events.APIGatewayProxyRequestContext{
			RequestID: "1234",
			Stage:     "prod",
		},
	}

	r, err := DefaultProxy(context.Background(), e)
	assert.NoError(t, err)

	assert.Equal(t, `example.com`, r.Host)
	assert.Equal(t, `prod`, r.Header.Get("X-Stage"))
	assert.Equal(t, `1234`, r.Header.Get("X-Request-Id"))
	assert.Equal(t, `18`, r.Header.Get("Content-Length"))
	assert.Equal(t, `application/json`, r.Header.Get("Content-Type"))
	assert.Equal(t, `bar`, r.Header.Get("X-Foo"))
}

func TestNewRequest_body(t *testing.T) {
	e := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Path:       "/pets",
		Body:       `{ "name": "Tobi" }`,
	}

	r, err := DefaultProxy(context.Background(), e)
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.Equal(t, `{ "name": "Tobi" }`, string(b))
}

func TestNewRequest_bodyBinary(t *testing.T) {
	e := events.APIGatewayProxyRequest{
		HTTPMethod:      "POST",
		Path:            "/pets",
		Body:            `aGVsbG8gd29ybGQK`,
		IsBase64Encoded: true,
	}

	r, err := DefaultProxy(context.Background(), e)
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.Equal(t, "hello world\n", string(b))
}

func customProxyStripPath(event events.APIGatewayProxyRequest, basePath string) (*http.Request, error) {
	r := NewRequest(context.TODO(), event)

	StripBasePath(basePath)(r)

	if err := r.CreateRequest(); err != nil {
		return nil, err
	}

	r.Transform(
		SetRemoteAddr,
		SetHeaderFields,
		SetContentLength,
		SetXRayHeader,
	)

	return r.Request, nil
}

func TestStripBasePath_executeapi(t *testing.T) {
	e := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Host": "xxxxxxxxxx.execute-api.us-east-1.amazonaws.com",
		},
		RequestContext: events.APIGatewayProxyRequestContext{
			Stage: "testing",
		},
	}

	t.Run("GetItem", func(t *testing.T) {
		e.Path = "/123"
		r, err := customProxyStripPath(e, "pets")
		assert.NoError(t, err)
		assert.Equal(t, "/123", r.URL.Path)
	})

	t.Run("ListItems", func(t *testing.T) {
		e.Path = "/"
		r, err := customProxyStripPath(e, "pets")
		assert.NoError(t, err)
		assert.Equal(t, "/", r.URL.Path)
	})
}

func TestStripBasePath_customDomain(t *testing.T) {
	e := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Host": "api.example.com",
		},
		RequestContext: events.APIGatewayProxyRequestContext{
			Stage: "testing",
		},
	}

	t.Run("ListItems", func(t *testing.T) {
		e.Path = "/pets"
		r, err := customProxyStripPath(e, "pets")
		assert.NoError(t, err)
		assert.Equal(t, "/", r.URL.Path)
	})

	t.Run("GetItem", func(t *testing.T) {
		e.Path = "/pets/123"
		r, err := customProxyStripPath(e, "pets")
		assert.NoError(t, err)
		assert.Equal(t, "/123", r.URL.Path)
	})
}

func TestStripBasePath_noBasePath(t *testing.T) {
	e := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Host": "api.example.com",
		},
		RequestContext: events.APIGatewayProxyRequestContext{
			Stage: "testing",
		},
	}

	t.Run("ListItems", func(t *testing.T) {
		e.Path = "/"
		r, err := customProxyStripPath(e, "")
		assert.NoError(t, err)
		assert.Equal(t, "/", r.URL.Path)
	})

	t.Run("GetItem", func(t *testing.T) {
		e.Path = "/123"
		r, err := customProxyStripPath(e, "")
		assert.NoError(t, err)
		assert.Equal(t, "/123", r.URL.Path)
	})
}
