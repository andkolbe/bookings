package render

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/andkolbe/bookings/internal/config"
	"github.com/andkolbe/bookings/internal/models"
	"github.com/justinas/nosurf"
)

// a FuncMap is a map of functions that can be used in a template
// Go allows us to create our own functions and pass them to the template
var functions = template.FuncMap{
	"humanDate": HumanDate,
	"formatDate": FormatDate,
	"iterate": Iterate,
	"add": Add,
}

var app *config.AppConfig
var pathToTemplates = "./templates"

func Add(a, b int) int {
	return a + b
}

// function that allows a user to iterate between two dates on the calendar
// returns a slice of ints, starting at 1, and going to count
func Iterate(count int) []int {
	var i int
	var items []int
	for i = 0; i < count; i++ {
		items = append(items, i)
	}
	return items
}

// sets the config for the template package
func NewRenderer(a *config.AppConfig) {
	app = a
}

// returns time in YYYY-MM-DD format to use in our templates
func HumanDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func FormatDate(t time.Time, f string) string {
	return t.Format(f)
}

func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	// flash messages appear once and then are automatically taken out of the session
	td.Flash = app.Session.PopString(r.Context(), "flash") // PopString puts something in our session and then removes it when we navigate away from the page
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning") 
	td.CSRFToken = nosurf.Token(r)
	// AddDefaultData has access to the session because it has the request
	if app.Session.Exists(r.Context(), "user_id") {
		td.IsAuthenticated = 1 // 1 means the user is logged in. 0 means the user is logged out
	}
	return td
}

// renders templates using html/template
func Template(w http.ResponseWriter, r *http.Request, html string, td *models.TemplateData) error {
	// in dev mode, don't use the template cache, instead rebuild it on every request
	var tc map[string]*template.Template

	if app.UseCache { // if use cache is true, use the template cache
		tc = app.TemplateCache
	} else { // else, rebuild a new template cache on every request
		tc, _ = CreateTemplateCache()
	}

	// create template cache when the app starts, then when we render a page, we are pulling a value from our config
	// get the template cache from the app config
	// tc := app.TemplateCache

	t, ok := tc[html] // if we get past this, then we have the template we want to use
	if !ok {
		return errors.New("Can't get template from cache")
	}

	buf := new(bytes.Buffer)

	td = AddDefaultData(td, r)

	_ = t.Execute(buf, td) // take the template we have, execute it, don't pass it any data and store the value in the buffer variable

	_, err := buf.WriteTo(w)
	if err != nil {
		fmt.Println("Error writing template to browser", err)
		return err
	}
	return nil

}

func CreateTemplateCache() (map[string]*template.Template, error) {
	// create a template cache that holds all our html templates in a map
	myCache := map[string]*template.Template{} // map with an index of type string and its contents are a pointer to template.Template

	// go to the templates folder, and get all of the files that start with anything but end with .page.html
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.html", pathToTemplates))
	if err != nil {
		return myCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err

		}

		// go to the templates folder, and get all of the files that end with .layout.html
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.html", pathToTemplates))
		if err != nil {
			return myCache, err

		}

		// if a .layout.html match is found, the length will be greater than 0
		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.html", pathToTemplates))
			if err != nil {
				return myCache, err

			}
		}
		myCache[name] = ts
	}
	return myCache, nil
}
