package apigo

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

type requestContextKey struct{}

var contextKey = &requestContextKey{}

// NewContext populates a context.Context from the http.Request with a
// request context provided in event from the AWS API Gateway proxy.
func NewContext(ctx context.Context, ev events.APIGatewayProxyRequest) context.Context {
	return context.WithValue(ctx, contextKey, ev.RequestContext)
}

// RequestContext returns the APIGatewayProxyRequestContext value stored in ctx.
func RequestContext(ctx context.Context) (events.APIGatewayProxyRequestContext, bool) {
	c, ok := ctx.Value(contextKey).(events.APIGatewayProxyRequestContext)
	return c, ok
}
