package apigo

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Gateway struct {
	Handler      http.Handler
	RequestProxy RequestProxy
}

// ListenAndServe is a drop-in replacement for http.ListenAndServe for use
// within AWS Lambda.
//
// NOTE: First, ignored argument to this function was ignored, because it was
// left to have the same signature of function, but also to behave as
// apex/gateway's ListenAndServe.
func ListenAndServe(_ string, h http.Handler) error {
	g := &Gateway{Handler: h}
	return g.ListenAndServe()
}

// ListenAndServe registers a listener of AWS Lambda events.
func (g *Gateway) ListenAndServe() error {
	if g.Handler == nil {
		g.Handler = http.DefaultServeMux
	}

	lambda.Start(g.Serve)

	return nil
}

// Serve handles incoming event from AWS Lambda by wraping them into
// http.Request which is further processed by http.Handler to reply
// as a APIGatewayProxyResponse.
func (g *Gateway) Serve(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if g.RequestProxy == nil {
		g.RequestProxy = NewRequest
	}

	r, err := g.RequestProxy(ctx, e)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	w := NewResponse()
	g.Handler.ServeHTTP(w, r)

	return w.End(), nil
}
