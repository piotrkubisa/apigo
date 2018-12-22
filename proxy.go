package apigo

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
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

type DefaultProxy struct {
	Host string
}

// DefaultProxy returns a new http.Request created from the given Lambda event.
func (p *DefaultProxy) Transform(ctx context.Context, ev events.APIGatewayProxyRequest) (*http.Request, error) {
	r := NewRequest(ctx, ev)
	r.Host = p.Host

	err := r.CreateRequest()
	if err != nil {
		return nil, err
	}

	AttachContext(r)
	SetRemoteAddr(r)
	SetHeaderFields(r)
	SetContentLength(r)
	SetCustomHeaders(r)
	SetXRayHeader(r)

	return r.Request, nil
}
