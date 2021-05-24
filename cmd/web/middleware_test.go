package main

import (
	"fmt"
	"net/http"
	"testing"
)


func TestNoSurf(t *testing.T) {
	var myH myHandler // type myHandler is defined in setup_test.go
	
	h := NoSurf(&myH) // we have to pass a http.Handler into NoSurf for it to run

	switch v := h.(type) {
	case http.Handler:
		// do nothing
	default:
		t.Error(fmt.Sprintf("type is not http.Handler, but is %T", v))
	}
}

func TestSessionLoad(t *testing.T) {
	var myH myHandler // type myHandler is defined in setup_test.go
	
	h := SessionLoad(&myH) // we have to pass a http.Handler into SessionLoad for it to run

	switch v := h.(type) {
	case http.Handler:
		// do nothing
	default:
		t.Error(fmt.Sprintf("type is not http.Handler, but is %T", v))
	}
}