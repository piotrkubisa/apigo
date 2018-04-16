package apigo

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/pkg/errors"
)

// RequestTransformer is a function which transforms an event and context
// provided from the API Gateway to http.Request.
type RequestTransformer func(context.Context, events.APIGatewayProxyRequest) (*http.Request, error)

// NewRequest returns a new http.Request created from the given Lambda event.
func NewRequest(ctx context.Context, ev events.APIGatewayProxyRequest) (*http.Request, error) {
	b := NewRequestBuilder(ctx, ev)
	err := b.Transform(b.DefaultTransforms()...)
	return b.Request, err
}

// RequestBuilder is an wrapper which helps transforming event from AWS API
// Gateway as a http.Request.
type RequestBuilder struct {
	ctx context.Context
	ev  events.APIGatewayProxyRequest

	Path    string
	URL     *url.URL
	Body    *strings.Reader
	Request *http.Request
}

// NewRequestBuilder defines new RequestBuilder with context and event data
// provided from the API Gateway.
func NewRequestBuilder(ctx context.Context, ev events.APIGatewayProxyRequest) *RequestBuilder {
	return &RequestBuilder{ctx: ctx, ev: ev}
}

// DefaultTransforms returns a collection of Transformer functions which all
// used by apigo during event-to-request transformation.
func (b *RequestBuilder) DefaultTransforms() []Transformer {
	return []Transformer{
		b.ParseURL,
		b.ParseBody,
		b.CreateRequest,
		b.AttachContext,
		b.SetRemoteAddr,
		b.SetHeaderFields,
		b.SetContentLength,
		b.SetCustomHeaders,
		b.SetXRayHeader,
	}
}

// Transformer transforms event from AWS API Gateway to the http.Request.
type Transformer func() error

// Transform AWS API Gateway event to a http.Request.
func (b *RequestBuilder) Transform(ts ...Transformer) error {
	if ts == nil || len(ts) == 0 {
		return errors.New("no transformers defined")
	}

	for _, p := range ts {
		if err := p(); err != nil {
			return err
		}
	}

	return nil
}

// ParseURL provides URL (as a *url.URL) to the RequestBuilder.
func (b *RequestBuilder) ParseURL() error {
	// Whether path has been already defined (i.e. processed by previous
	// function) then use it, otherwise use path from the event.
	path := b.Path
	if len(path) == 0 {
		path = b.ev.Path
	}

	// Parse URL to *url.URL
	u, err := url.Parse(path)
	if err != nil {
		return errors.Wrap(err, "parsing path")
	}

	// Query-string
	q := u.Query()
	for k, v := range b.ev.QueryStringParameters {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	// Host
	u.Host = b.ev.Headers["Host"]

	b.URL = u
	return nil
}

// ParseBody provides body of the request to the RequestBuilder.
func (b *RequestBuilder) ParseBody() error {
	body := b.ev.Body
	if b.ev.IsBase64Encoded {
		b, err := base64.StdEncoding.DecodeString(body)
		if err != nil {
			return errors.Wrap(err, "decoding base64 body")
		}
		body = string(b)
	}

	b.Body = strings.NewReader(body)
	return nil
}

// CreateRequest provides *http.Request to the RequestBuilder.
func (b *RequestBuilder) CreateRequest() error {
	req, err := http.NewRequest(b.ev.HTTPMethod, b.URL.String(), b.Body)
	if err != nil {
		return errors.Wrap(err, "creating request")
	}

	b.Request = req
	return nil
}

// AttachContext attaches events' RequestContext to the http.Request.
func (b *RequestBuilder) AttachContext() error {
	b.Request = b.Request.WithContext(NewContext(b.Request.Context(), b.ev))
	return nil
}

// SetRemoteAddr sets RemoteAddr to the request.
func (b *RequestBuilder) SetRemoteAddr() error {
	b.Request.RemoteAddr = b.ev.RequestContext.Identity.SourceIP
	return nil
}

// SetHeaderFields sets headers to the request.
func (b *RequestBuilder) SetHeaderFields() error {
	for k, v := range b.ev.Headers {
		b.Request.Header.Set(k, v)
	}
	return nil
}

// SetContentLength sets Content-Length to the request if it has not been set.
func (b *RequestBuilder) SetContentLength() error {
	if b.Request.Header.Get("Content-Length") == "" {
		b.Request.Header.Set("Content-Length", strconv.Itoa(b.Body.Len()))
	}
	return nil
}

// SetCustomHeaders sets additional headers from the event's RequestContext to
// the request.
func (b *RequestBuilder) SetCustomHeaders() error {
	b.Request.Header.Set("X-Request-Id", b.ev.RequestContext.RequestID)
	b.Request.Header.Set("X-Stage", b.ev.RequestContext.Stage)
	return nil
}

// SetXRayHeader sets AWS X-Ray Trace ID from the event's context.
func (b *RequestBuilder) SetXRayHeader() error {
	if traceID := b.ctx.Value("x-amzn-trace-id"); traceID != nil {
		b.Request.Header.Set("X-Amzn-Trace-Id", fmt.Sprintf("%v", traceID))
	}
	return nil
}
