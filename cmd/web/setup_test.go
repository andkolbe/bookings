// whatever is written in here will run before our tests run

package main

import (
	"net/http"
	"os"
	"testing"
)


func TestMain(m *testing.M) { // there must be a func called TestMain in the setup_test.go file


	os.Exit(m.Run()) // exit all of the tests when they are finished
}

// set up an http.Handler that we can use in our testing environment only
type myHandler struct{}
func (mh *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
