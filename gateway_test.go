package apigo_test

import (
	"fmt"
	"log"
	"net/http"

	"github.com/piotrkubisa/apigo"
)

func Example() {
	http.HandleFunc("/", hello)
	log.Fatal(apigo.ListenAndServe(":3000", nil))
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello World from Go")
}
