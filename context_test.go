package apigo

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/tj/assert"
)

func TestNewContext(t *testing.T) {
	ev := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			AccountID:  "000000000000",
			ResourceID: "XXXXXX",
			Stage:      "testing",
			RequestID:  "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX",
			Identity: events.APIGatewayRequestIdentity{
				CognitoIdentityPoolID:         "xxx",
				AccountID:                     "xxx",
				CognitoIdentityID:             "xxx",
				Caller:                        "xxx",
				APIKey:                        "xxx",
				SourceIP:                      "1.1.1.1",
				CognitoAuthenticationType:     "xxx",
				CognitoAuthenticationProvider: "xxx",
				UserArn:   "xxx",
				UserAgent: "xxx",
				User:      "xxx",
			},
			ResourcePath: "/{id}",
			Authorizer: map[string]interface{}{
				"cognitoUsername": "johndoe",
				"principalId":     "xxxx",
			},
			HTTPMethod: "GET",
			APIID:      "XXXXXXXXXX",
		},
	}

	r, _ := DefaultProxy(context.TODO(), ev)
	rc, _ := RequestContext(r.Context())

	assert.Equal(t, "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX", rc.RequestID)
	assert.Equal(t, "johndoe", rc.Authorizer["cognitoUsername"])
}

func TestNewRequest_context(t *testing.T) {
	type testRequestContextKey struct{}
	var testContextKey = &testRequestContextKey{}

	ev := events.APIGatewayProxyRequest{}
	ctx := context.WithValue(context.TODO(), testContextKey, "value")

	r, err := DefaultProxy(ctx, ev)
	assert.NoError(t, err)

	v := r.Context().Value(testContextKey)
	assert.Equal(t, "value", v)
}
