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

type RequestProxy func(context.Context, events.APIGatewayProxyRequest) (*http.Request, error)

// NewRequest returns a new http.Request from the given Lambda event.
func NewRequest(ctx context.Context, ev events.APIGatewayProxyRequest) (*http.Request, error) {
	ep := NewEventRequestParser(ctx, ev)
	err := ep.Parse(
		ep.ParseURL,
		ep.ParseBody,
		ep.CreateRequest,
		ep.SetRemoteAddr,
		ep.SetHeaderFields,
		ep.SetContentLength,
		ep.SetCustomHeaders,
		ep.SetXRayHeader,
	)
	return ep.Request, err
}

type EventRequestParser struct {
	ctx context.Context
	ev  events.APIGatewayProxyRequest

	Path    string
	URL     *url.URL
	Body    *strings.Reader
	Request *http.Request
}

func NewEventRequestParser(ctx context.Context, ev events.APIGatewayProxyRequest) *EventRequestParser {
	return &EventRequestParser{ctx: ctx, ev: ev}
}

func (ep *EventRequestParser) Parse(parsers ...Parser) error {
	if parsers == nil || len(parsers) == 0 {
		return errors.New("no parsers defined")
	}

	for _, p := range parsers {
		if err := p(); err != nil {
			return err
		}
	}

	return nil
}

type Parser func() error

func (ep *EventRequestParser) ParseURL() error {
	path := ep.Path
	if len(path) == 0 {
		path = ep.ev.Path
	}

	// path
	u, err := url.Parse(path)
	if err != nil {
		return errors.Wrap(err, "parsing path")
	}

	// querystring
	q := u.Query()
	for k, v := range ep.ev.QueryStringParameters {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	u.Host = ep.ev.Headers["Host"]

	ep.URL = u
	return nil
}

func (ep *EventRequestParser) ParseBody() error {
	body := ep.ev.Body
	if ep.ev.IsBase64Encoded {
		b, err := base64.StdEncoding.DecodeString(body)
		if err != nil {
			return errors.Wrap(err, "decoding base64 body")
		}
		body = string(b)
	}

	ep.Body = strings.NewReader(body)
	return nil
}

func (ep *EventRequestParser) CreateRequest() error {
	req, err := http.NewRequest(ep.ev.HTTPMethod, ep.URL.String(), ep.Body)
	if err != nil {
		return errors.Wrap(err, "creating request")
	}

	ep.Request = req
	return nil
}

func (ep *EventRequestParser) SetRemoteAddr() error {
	ep.Request.RemoteAddr = ep.ev.RequestContext.Identity.SourceIP
	return nil
}

func (ep *EventRequestParser) SetHeaderFields() error {
	for k, v := range ep.ev.Headers {
		ep.Request.Header.Set(k, v)
	}
	return nil
}

func (ep *EventRequestParser) SetContentLength() error {
	if ep.Request.Header.Get("Content-Length") == "" {
		ep.Request.Header.Set("Content-Length", strconv.Itoa(ep.Body.Len()))
	}
	return nil
}

func (ep *EventRequestParser) SetCustomHeaders() error {
	ep.Request.Header.Set("X-Request-Id", ep.ev.RequestContext.RequestID)
	ep.Request.Header.Set("X-Stage", ep.ev.RequestContext.Stage)
	return nil
}

func (ep *EventRequestParser) SetXRayHeader() error {
	if traceID := ep.ctx.Value("x-amzn-trace-id"); traceID != nil {
		ep.Request.Header.Set("X-Amzn-Trace-Id", fmt.Sprintf("%v", traceID))
	}
	return nil
}
