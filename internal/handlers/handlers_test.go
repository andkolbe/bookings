package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

//
type postData struct {
	key   string // name of the form input
	value string // what is typed in the form input
}

var theTests = []struct {
	name   string
	url    string
	method string
	params []postData
	expectedStatusCode int
}{
	{"home", "/", "GET", []postData{}, http.StatusOK},
	{"about", "/about", "GET", []postData{}, http.StatusOK},
	{"gq", "/generals-quarters", "GET", []postData{}, http.StatusOK},
	{"ms", "/majors-suite", "GET", []postData{}, http.StatusOK},
	{"sa", "/search-availability", "GET", []postData{}, http.StatusOK},
	{"contact", "/contact", "GET", []postData{}, http.StatusOK},
	{"mr", "/make-reservation", "GET", []postData{}, http.StatusOK},
	{"post-search-avail", "/search-availability", "POST", []postData{
		{key: "start", value: "2020-01-01"},
		{key: "end", value: "2020-01-02"},
	}, http.StatusOK},
	{"post-search-avail-json", "/search-availability-json", "POST", []postData{
		{key: "start", value: "2020-01-01"},
		{key: "end", value: "2020-01-02"},
	}, http.StatusOK},
	{"make reservation post", "/make-reservation", "POST", []postData{
		{key: "first_name", value: "Andrew"},
		{key: "last_name", value: "Kolbe"},
		{key: "email", value: "andrew@andrew.com"},
		{key: "phone", value: "111-111-1111"},
	}, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close() // what is written after defer doesn't get executed until after the current function is finished

	for _, e := range theTests {
		if e.method == "GET" {
			resp, err := ts.Client().Get(ts.URL + e.url) // tell the client to make a GET request to the url that we want to test
			if err != nil {
				t.Log(err) 
				t.Fatal(err)
			}

			if resp.StatusCode != e.expectedStatusCode {
				t.Errorf("for %s, expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
			}
		} else { // POST
			values := url.Values{} // url.Values is part of the standard library that holds information as a POST request variable
			// loop over all of the params in our POST request
			for _, x := range e.params {
				values.Add(x.key, x.value)
			}
			resp, err := ts.Client().PostForm(ts.URL + e.url, values)
			if err != nil {
				t.Log(err) 
				t.Fatal(err)
			}

			if resp.StatusCode != e.expectedStatusCode {
				t.Errorf("for %s, expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
			}
		}
	}
}
