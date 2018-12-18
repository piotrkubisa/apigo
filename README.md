# apigo

[![Documentation](https://godoc.org/github.com/piotrkubisa/apigo?status.svg)](http://godoc.org/github.com/piotrkubisa/apigo)
[![Build Status](https://travis-ci.org/piotrkubisa/apigo.svg?branch=master)](https://travis-ci.org/piotrkubisa/apigo)
[![Go Report Card](https://goreportcard.com/badge/github.com/piotrkubisa/apigo)](https://goreportcard.com/report/github.com/piotrkubisa/apigo)

Package `apigo` is an drop-in adapter to AWS Lambda functions (based on `go1.x` runtime) with a AWS API Gateway to easily reuse logic from _serverfull_ `http.Handler`s and provide the same experience for serverless function.

## Installation

Add `apigo` dependency using your vendor package manager (i.e. `dep`) or `go get` it:

```bash
go get -v github.com/piotrkubisa/apigo
```

## Usage

### Default behaviour

If you have already registered some `http.Handler`s, you can easily reuse them with `apigo.Gateway`.
Example below illustrates how to create a _hello world_ serverless application with `apigo`:

```go
package main

import (
	"net/http"

	"github.com/piotrkubisa/apigo"
)

func main() {
	http.HandleFunc("/hello", helloHandler)

	apigo.ListenAndServe(http.DefaultServeMux)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`"Hello World"`))
}
```

### Custom event-to-request transformation

If you have a bit more sophisticated deployment of your AWS Lambda functions then you probably would love to have more control over _event-to-request_ transformation.
Imagine a situation if you have your API in one serverless function and you also have additional [custom authorizer](https://aws.amazon.com/blogs/compute/introducing-custom-authorizers-in-amazon-api-gateway/) in separate AWS Lamda function.
In following scenario (presented in example below) context variable provided by serverless authorizer is passed to the API's `http.Request` context, which can be further inspected during request handling:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/go-chi/chi"
	"github.com/piotrkubisa/apigo"
)

func main() {
	g := &apigo.Gateway{
		Proxy:   apigo.ProxyFunc(customProxy),
		Handler: routing(),
	}
	g.ListenAndServe()
}

type contextUsername struct{}

var keyUsername = &contextUsername{}

func customProxy(ctx context.Context, ev events.APIGatewayProxyRequest) (*http.Request, error) {
	r, err := apigo.DefaultProxy(ctx, ev)
	if err != nil {
		return nil, err
	}
	// Add username to the http.Request's context from the custom authorizer
	r = r.WithContext(
		context.WithValue(
			r.Context(),
			keyUsername,
			ev.RequestContext.Authorizer["username"],
		),
	)
	return r, err
}

func routing() http.Handler {
	r := chi.NewRouter()

	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("id: %s", chi.URLParam(r, "id"))

		// Remember: headers, status and then payload - always in this order
		// set headers
		w.Header().Set("Content-Type", "application/json")
		// set status
		w.WriteHeader(http.StatusOK)
		// set response payload
		username, _ := r.Context().Value(keyUsername).(string)
		fmt.Fprintf(w, `"Hello %s"`, username)
	})

	return r
}
```

## Goroutines

If you are going to use `goroutines` in your AWS handler, then it is worth noting you should control its execution (i.e. by using `sync.WaitGroup`), otherwise code in the `goroutine` might be killed after returning a response to AWS API Gateway.

```go
package main

import (
	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"github.com/piotrkubisa/apigo"
)

func main() {
	apigo.ListenAndServe(routing())
}

func routing() http.Handler {
	r := chi.NewRouter()

	r.Post("/{name}", func(w http.ResponseWriter, r *http.Request) {
		var wg sync.WaitGroup
		wg.Add(2)
		go sendIoTMessage(&wg)
		go sendSlackNotification(&wg)
		wg.Wait()

		// Headers, status, payload
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`"Done"`))
	})

	return r
}

func sendIoTMessage(wg *sync.WaitGroup) {
	// ...
	wg.Done()
}

func sendSlackNotification(wg *sync.WaitGroup) {
	// ...
	wg.Done()
}
```

## Credits

Project has been forked from fabulous [tj's](https://github.com/tj) [apex/gateway](https://github.com/apex) repository,
at [0bee09a](https://github.com/piotrkubisa/apigo/commit/0bee09ab83e1d4ea098e77c38ce90890a25c42cb).
