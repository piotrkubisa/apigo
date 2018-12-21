package apigo_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
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

	payload, err := ioutil.ReadFile("./vendor/github.com/aws/aws-lambda-go/events/testdata/apigw-request.json")
	if err != nil {
		panic(err)
	}
	var ev events.APIGatewayProxyRequest
	if err := json.Unmarshal(payload, &ev); err != nil {
		panic(err)
	}

	for i := 0; i < b.N; i++ {
		g.Serve(context.TODO(), ev)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTeapot)
	w.Write([]byte(`"Hello World"`))
}
