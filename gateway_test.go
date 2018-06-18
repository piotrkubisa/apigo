package apigo_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/piotrkubisa/apigo"
)

func BenchmarkGateway_Serve(b *testing.B) {
	g := apigo.Gateway{
		Handler: http.HandlerFunc(helloHandler),
		Proxy:   apigo.ProxyFunc(apigo.DefaultProxy),
	}

	sampleEvent := events.APIGatewayProxyRequest{
		Resource:   "/{proxy+}",
		Path:       "/_version",
		HTTPMethod: "GET",
		Headers: map[string]string{
			"Accept":                       "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			"Accept-Encoding":              "gzip, deflate, br",
			"Accept-Language":              "en-US,en;q=0.9",
			"CloudFront-Forwarded-Proto":   "https",
			"CloudFront-Is-Desktop-Viewer": "true",
			"CloudFront-Is-Mobile-Viewer":  "false",
			"CloudFront-Is-SmartTV-Viewer": "false",
			"CloudFront-Is-Tablet-Viewer":  "false",
			"CloudFront-Viewer-Country":    "DE",
			"Host":              "tester1337.execute-api.eu-west-1.amazonaws.com",
			"User-Agent":        "Undefined/1.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Safari/537.36",
			"Via":               "2.0 cccccccccccccccccccccccccccccccc.cloudfront.net (CloudFront)",
			"X-Amz-Cf-Id":       "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee_eeeeeeeeeeeee==",
			"X-Amzn-Trace-Id":   "Root=1-aaaaaaaa-bbbbbbbbbbbbbbbbbbbbbbbb",
			"X-Forwarded-For":   "0.1.2.225, 0.137.255.255",
			"X-Forwarded-Port":  "443",
			"X-Forwarded-Proto": "https",
			"dnt":               "1",
			"upgrade-insecure-requests": "1",
		},
		QueryStringParameters: nil,
		PathParameters: map[string]string{
			"proxy": "hello",
		},
		StageVariables: nil,
		RequestContext: events.APIGatewayProxyRequestContext{
			AccountID:  "123456789012",
			ResourceID: "abc123",
			Stage:      "prod",
			RequestID:  "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			Identity: events.APIGatewayRequestIdentity{
				CognitoIdentityPoolID:         "",
				AccountID:                     "",
				CognitoIdentityID:             "",
				Caller:                        "",
				APIKey:                        "",
				SourceIP:                      "0.1.2.225",
				CognitoAuthenticationType:     "",
				CognitoAuthenticationProvider: "",
				UserArn:   "",
				UserAgent: "Undefined/1.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Safari/537.36",
				User:      "",
			},
			ResourcePath: "/{proxy+}",
			Authorizer:   nil,
			HTTPMethod:   "GET",
			APIID:        "tester1337",
		},
		Body: "",
	}

	for i := 0; i < b.N; i++ {
		g.Serve(context.TODO(), sampleEvent)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTeapot)
	w.Write([]byte(`"Hello World"`))
}
