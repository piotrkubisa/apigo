package apigo

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Gateway mimics the http.Server definition and takes care of proxying
// AWS Lambda event to http.Request via Proxy and then handling it using Handler
type Gateway struct {
	Proxy   Proxy
	Handler http.Handler
}

// NewGateway creates new Gateway, which utilizes handler
// (or http.DefaultServeMux if nil passed) as a Gateway.Handler and
// apigo.http.DefaultProxy as a Gateway.Proxy.
func NewGateway(host string, handler http.Handler) *Gateway {
	if handler == nil {
		handler = http.DefaultServeMux
	}

	return &Gateway{
		Handler: handler,
		Proxy:   &DefaultProxy{host},
	}
}

// ListenAndServe is a drop-in replacement for http.ListenAndServe for use
// within AWS Lambda.
func ListenAndServe(host string, h http.Handler) {
	NewGateway(host, h).ListenAndServe()
}

// ListenAndServe registers a listener of AWS Lambda events.
func (g *Gateway) ListenAndServe() {
	lambda.Start(g.Serve)
}

// Serve handles incoming event from AWS Lambda by wraping them into
// http.Request which is further processed by http.Handler to reply
// as a APIGatewayProxyResponse.
func (g *Gateway) Serve(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	r, err := g.Proxy.Transform(ctx, e)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	w := NewResponse()
	g.Handler.ServeHTTP(w, r)

	return w.End(), nil
}
