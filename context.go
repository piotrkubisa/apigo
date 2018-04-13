package apigo

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

type contextKey int

const (
	requestContextKey contextKey = iota
)

// AttachRequestContext populates a context.Context from the http.Request with a
// request context provided in event from the AWS API Gateway proxy.
func AttachRequestContext(r *http.Request, ev events.APIGatewayProxyRequest) *http.Request {
	return r.WithContext(
		context.WithValue(r.Context(), requestContextKey, ev.RequestContext),
	)
}

// RequestContext returns the APIGatewayProxyRequestContext value stored in ctx.
func RequestContext(ctx context.Context) (events.APIGatewayProxyRequestContext, bool) {
	c, ok := ctx.Value(requestContextKey).(events.APIGatewayProxyRequestContext)
	return c, ok
}
