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
				CognitoIdentityPoolID:         "",
				AccountID:                     "",
				CognitoIdentityID:             "",
				Caller:                        "",
				APIKey:                        "",
				SourceIP:                      "1.1.1.1",
				CognitoAuthenticationType:     "",
				CognitoAuthenticationProvider: "",
				UserArn:   "",
				UserAgent: "PostmanRuntime/7.1.1",
				User:      "",
			},
			ResourcePath: "/{id}",
			Authorizer: map[string]interface{}{
				"cognitoUsername": "johndoe",
				"principalId":     "XXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX",
			},
			HTTPMethod: "GET",
			APIID:      "XXXXXXXXXX",
		},
	}

	r, _ := NewRequest(context.TODO(), ev)
	rc, _ := RequestContext(r.Context())

	assert.Equal(t, "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX", rc.RequestID)
	assert.Equal(t, "johndoe", rc.Authorizer["cognitoUsername"])
}
