package apigo

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestNewRequest_path(t *testing.T) {
	e := events.APIGatewayProxyRequest{
		Path: "/pets/luna",
	}

	r, err := new(DefaultProxy).Transform(context.TODO(), e)
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

	r, err := new(DefaultProxy).Transform(context.TODO(), e)
	assert.NoError(t, err)

	assert.Equal(t, "DELETE", r.Method)
}

func TestNewRequest_queryString(t *testing.T) {
	e := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/pets",
		MultiValueQueryStringParameters: map[string][]string{
			"order":  {"desc"},
			"fields": {"name,species"},
		},
	}

	r, err := new(DefaultProxy).Transform(context.TODO(), e)
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

	r, err := new(DefaultProxy).Transform(context.TODO(), e)
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
		},
		MultiValueHeaders: map[string][]string{
			"Content-Type": {"application/json"},
			"X-Foo":        {"bar"},
		},
		RequestContext: events.APIGatewayProxyRequestContext{
			RequestID: "1234",
			Stage:     "prod",
		},
	}

	p := &DefaultProxy{Host: "api.example.com"}
	r, err := p.Transform(context.TODO(), e)
	assert.NoError(t, err)

	assert.Equal(t, `api.example.com`, r.Host)
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

	r, err := new(DefaultProxy).Transform(context.TODO(), e)
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

	r, err := new(DefaultProxy).Transform(context.TODO(), e)
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.Equal(t, "hello world\n", string(b))
}

func TestStripBasePath_executeapi(t *testing.T) {
	p := StripBasePathProxy{
		Host:     "xxxxxxxxxx.execute-api.us-east-1.amazonaws.com",
		BasePath: "pets",
	}

	e := events.APIGatewayProxyRequest{
		MultiValueHeaders: map[string][]string{
			"Host": {p.Host},
		},
		RequestContext: events.APIGatewayProxyRequestContext{
			Stage: "testing",
		},
	}

	t.Run("GetItem", func(t *testing.T) {
		e.Path = "/123"
		r, err := p.Transform(context.TODO(), e)
		assert.NoError(t, err)
		assert.Equal(t, "/123", r.URL.Path)
	})

	t.Run("ListItems", func(t *testing.T) {
		e.Path = "/"
		r, err := p.Transform(context.TODO(), e)
		assert.NoError(t, err)
		assert.Equal(t, "/", r.URL.Path)
	})
}

func TestStripBasePath_customDomain(t *testing.T) {
	p := StripBasePathProxy{
		Host:     "api.example.com",
		BasePath: "pets",
	}

	e := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Host": p.Host,
		},
		MultiValueHeaders: map[string][]string{
			"Host": {p.Host},
		},
		RequestContext: events.APIGatewayProxyRequestContext{
			Stage: "testing",
		},
	}

	t.Run("ListItems", func(t *testing.T) {
		e.Path = "/pets"
		r, err := p.Transform(context.TODO(), e)
		assert.NoError(t, err)
		assert.Equal(t, "/", r.URL.Path)
	})

	t.Run("GetItem", func(t *testing.T) {
		e.Path = "/pets/123"
		r, err := p.Transform(context.TODO(), e)
		assert.NoError(t, err)
		assert.Equal(t, "/123", r.URL.Path)
	})
}

func TestStripBasePath_noBasePath(t *testing.T) {
	p := StripBasePathProxy{
		Host:     "api.example.com",
		BasePath: "pets",
	}

	e := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Host": p.Host,
		},
		MultiValueHeaders: map[string][]string{
			"Host": {p.Host},
		},
		RequestContext: events.APIGatewayProxyRequestContext{
			Stage: "testing",
		},
	}

	t.Run("ListItems", func(t *testing.T) {
		e.Path = "/"
		r, err := p.Transform(context.TODO(), e)
		assert.NoError(t, err)
		assert.Equal(t, "/", r.URL.Path)
	})

	t.Run("GetItem", func(t *testing.T) {
		e.Path = "/123"
		r, err := p.Transform(context.TODO(), e)
		assert.NoError(t, err)
		assert.Equal(t, "/123", r.URL.Path)
	})
}
