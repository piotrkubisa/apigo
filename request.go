package apigo

import (
	"bytes"
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

// Request is an wrapper which helps transforming event from AWS API
// Gateway as a http.Request.
type Request struct {
	Context context.Context
	Event   events.APIGatewayProxyRequest

	Path    string
	Host    string
	URL     *url.URL
	Body    *bytes.Reader
	Request *http.Request
}

// NewRequest defines new RequestBuilder with context and event data
// provided from the API Gateway.
func NewRequest(ctx context.Context, ev events.APIGatewayProxyRequest) *Request {
	return &Request{
		Context: ctx,
		Event:   ev,
	}
}

// StripBasePath removes a BasePath from the Path fragment of the URL.
// StripBasePath must be run before RequestBuilder.ParseURL function.
func (r *Request) StripBasePath(basePath string) {
	r.Path = omitBasePath(r.Event.Path, basePath)
}

// omitBasePath strips out the base path from the given path.
//
// It allows to support both API endpoints (default, auto-generated
// "execute-api" address and configured Base Path Mapping/ with a Custom Domain
// Name), while preserving the same routing registered on the http.Handler.
func omitBasePath(path string, basePath string) string {
	if path == "/" || basePath == "" {
		return path
	}

	if strings.HasPrefix(path, "/"+basePath) {
		path = strings.Replace(path, basePath, "", 1)
	}
	if strings.HasPrefix(path, "//") {
		path = path[1:]
	}

	return path
}

// CreateRequest provides *http.Request to the RequestBuilder.
func (r *Request) CreateRequest() error {
	u, err := r.ParseURL()
	if err != nil {
		return err
	}
	if err := r.ParseBody(); err != nil {
		return err
	}

	req, err := http.NewRequest(r.Event.HTTPMethod, u.String(), r.Body)
	if err != nil {
		return errors.Wrap(err, "creating request")
	}

	r.Request = req
	return nil
}

// ParseURL provides URL (as a *url.URL) to the RequestBuilder.
func (r *Request) ParseURL() (*url.URL, error) {
	// Whether path has been already defined (i.e. processed by previous
	// function) then use it, otherwise use path from the event.
	path := r.Path
	if len(path) == 0 {
		path = r.Event.Path
	}

	// Parse URL to *url.URL
	u, err := url.Parse(path)
	if err != nil {
		return nil, errors.Wrap(err, "parsing path")
	}

	// Query-string
	q := url.Values(r.Event.MultiValueQueryStringParameters)
	u.RawQuery = q.Encode()

	// Host
	u.Host = r.Host

	return u, nil
}

// ParseBody provides body of the request to the RequestBuilder.
func (r *Request) ParseBody() error {
	body := []byte(r.Event.Body)
	if r.Event.IsBase64Encoded {
		b, err := base64.StdEncoding.DecodeString(r.Event.Body)
		if err != nil {
			return errors.Wrap(err, "decoding base64 body")
		}
		body = b
	}
	r.Body = bytes.NewReader(body)
	return nil
}

// AttachContext attaches events' RequestContext to the http.Request.
func (r *Request) AttachContext() {
	r.Request = r.Request.WithContext(NewContext(r.Context, r.Event))
}

// SetRemoteAddr sets RemoteAddr to the request.
func (r *Request) SetRemoteAddr() {
	r.Request.RemoteAddr = r.Event.RequestContext.Identity.SourceIP
}

// SetHeaderFields sets headers to the request.
func (r *Request) SetHeaderFields() {
	for k, hs := range r.Event.MultiValueHeaders {
		for _, v := range hs {
			r.Request.Header.Add(k, v)
		}
	}
}

// SetContentLength sets Content-Length to the request if it has not been set.
func (r *Request) SetContentLength() {
	if r.Request.Header.Get("Content-Length") == "" {
		r.Request.Header.Set("Content-Length", strconv.Itoa(r.Body.Len()))
	}
}

// SetCustomHeaders assigns X-Request-Id and X-Stage from the event's
// Request Context.
func (r *Request) SetCustomHeaders() {
	r.Request.Header.Set("X-Request-Id", r.Event.RequestContext.RequestID)
	r.Request.Header.Set("X-Stage", r.Event.RequestContext.Stage)
}

// SetXRayHeader sets AWS X-Ray Trace ID from the event's context.
func (r *Request) SetXRayHeader() {
	if traceID := r.Context.Value("x-amzn-trace-id"); traceID != nil {
		r.Request.Header.Set("X-Amzn-Trace-Id", fmt.Sprintf("%v", traceID))
	}
}
