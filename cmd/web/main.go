// The first thing that must appear in a go file is what package you are using
// you can name it whatever you want, but the standard is to call it main
package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/andkolbe/bookings/internal/config"
	"github.com/andkolbe/bookings/internal/handlers"
	"github.com/andkolbe/bookings/internal/models"
	"github.com/andkolbe/bookings/internal/render"
)

const portNumber = ":8080"

var app config.AppConfig

var session *scs.SessionManager

// main is the main application function
func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(fmt.Sprintf("Starting application on port %s", portNumber))
	// _ = http.ListenAndServe(portNumber, nil)

	srv := &http.Server{
		Addr: portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() error {
	// what I am going to put in the session
	gob.Register(models.Reservation{})

	// change this to true when in production
	app.InProduction = false

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true // true = cookie persists even if the browser window closes
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
		return err
	}

	app.TemplateCache = tc
	app.UseCache  = false

	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	// gives the render component of our app access to the app config variable
	render.NewTemplates(&app)

	return nil
}