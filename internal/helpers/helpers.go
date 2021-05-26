package helpers

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/andkolbe/bookings/internal/config"
)


var app *config.AppConfig

// sets up app config for helpers
func NewHelpers(a *config.AppConfig) {
	app = a
}

func ClientError(w http.ResponseWriter, status int) {
	// because it is a client error, we need to show the client a response 
	app.InfoLog.Println("Client error with status of", status)
	http.Error(w, http.StatusText(status), status)
}

func ServerError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.ErrorLog.Println(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func IsAuthenticated(r *http.Request) bool { // return true or false if they are authenticated or not
	exists := app.Session.Exists(r.Context(), "user_id")
	return exists
}