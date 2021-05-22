package render

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/andkolbe/bookings/internal/config"
	"github.com/andkolbe/bookings/internal/models"
	"github.com/justinas/nosurf"
)

// a FuncMap is a map of functions that can be used in a template
// Go allows us to create our own functions and pass them to the template
var functions = template.FuncMap{}

var app *config.AppConfig

// sets the config for the template package
func NewTemplates(a *config.AppConfig) {
	app = a
}

func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.CSRFToken = nosurf.Token(r)
	return td
}

// renders templates using html/template
func RenderTemplate(w http.ResponseWriter, r *http.Request, html string, td *models.TemplateData) {
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
		log.Fatal("Could not get template from template cache")
	}

	buf := new(bytes.Buffer)

	td = AddDefaultData(td, r)

	_ = t.Execute(buf, td) // take the template we have, execute it, don't pass it any data and store the value in the buffer variable

	_, err := buf.WriteTo(w)
	if err != nil {
		fmt.Println("Error writing template to browser", err)
	}

}

func CreateTemplateCache() (map[string]*template.Template, error) {
	// create a template cache that holds all our html templates in a map
	myCache := map[string]*template.Template{} // map with an index of type string and its contents are a pointer to template.Template

	// go to the templates folder, and get all of the files that start with anything but end with .page.html
	pages, err := filepath.Glob("./templates/*.page.html")
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
		matches, err := filepath.Glob("./templates/*.layout.html")
		if err != nil {
			return myCache, err

		}

		// if a .layout.html match is found, the length will be greater than 0
		if len(matches) > 0 {
			ts, err = ts.ParseGlob("./templates/*.layout.html")
			if err != nil {
				return myCache, err

			}
		}
		myCache[name] = ts
	}
	return myCache, nil
}
