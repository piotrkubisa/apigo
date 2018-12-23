package apigo

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/pkg/errors"
)

// Proxy transforms an event and context provided from the API Gateway
// to the http.Request.
type Proxy interface {
	Transform(context.Context, events.APIGatewayProxyRequest) (*http.Request, error)
}

// ProxyFunc implements the Proxy interface to allow use of ordinary function
// as a handler.
type ProxyFunc func(context.Context, events.APIGatewayProxyRequest) (*http.Request, error)

// Transform calls f(ctx, ev).
func (f ProxyFunc) Transform(ctx context.Context, ev events.APIGatewayProxyRequest) (*http.Request, error) {
	return f(ctx, ev)
}

// DefaultProxy is a default proxy for AWS API Gateway events
type DefaultProxy struct {
	Host string
}

// Transform returns a new http.Request created from the given Lambda event.
func (p *DefaultProxy) Transform(ctx context.Context, ev events.APIGatewayProxyRequest) (*http.Request, error) {
	r := NewRequest(ctx, ev)

	req, err := r.CreateRequest(p.Host)
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}

	r.AttachContext(req)
	r.SetRemoteAddr(req)
	r.SetHeaderFields(req)
	r.SetContentLength(req)
	r.SetCustomHeaders(req)
	r.SetXRayHeader(req)

	return req, nil
}

type StripBasePathProxy struct {
	Host     string
	BasePath string
}

func (p *StripBasePathProxy) Transform(ctx context.Context, ev events.APIGatewayProxyRequest) (*http.Request, error) {
	r := NewRequest(ctx, ev)
	r.StripBasePath(p.BasePath)

	req, err := r.CreateRequest(p.Host)
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}

	r.AttachContext(req)
	r.SetRemoteAddr(req)
	r.SetHeaderFields(req)
	r.SetContentLength(req)
	r.SetCustomHeaders(req)
	r.SetXRayHeader(req)

	return req, nil
}
