package apigo_test

import (
	"log"
	"net/http"

	"github.com/piotrkubisa/apigo"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTeapot)
	w.Write([]byte(`"Hello World"`))
}

func Example() {
	http.HandleFunc("/", helloHandler)
	log.Fatal(apigo.ListenAndServe("", nil))
}
