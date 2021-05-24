package render

import (
	"encoding/gob"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/andkolbe/bookings/internal/config"
	"github.com/andkolbe/bookings/internal/models"
)


var session *scs.SessionManager
var testApp config.AppConfig

func TestMain(m *testing.M) {

	// what I am going to put in the session
	gob.Register(models.Reservation{})

	// change this to true when in production
	testApp.InProduction = false

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true // true = cookie persists even if the browser window closes
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = false

	testApp.Session = session

	app = &testApp // makes sure that app (defined inside of render.go) is populated with the testApp data

	os.Exit(m.Run())
}

// create an interface that satisfies the requirements for a response writer
type myWriter struct{}

func (tw *myWriter) Header() http.Header {
	var h http.Header
	return h
}

func (tw *myWriter) WriteHeader(i int) {

}

func (tw  *myWriter) Write(b []byte) (int, error) {
	length := len(b)
	return length, nil
}