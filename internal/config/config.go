package config

import (
	"html/template"
	"log"

	"github.com/alexedwards/scs/v2"
	"github.com/andkolbe/bookings/internal/models"
)

// holds the application config
// because it is a struct, we can put anything we need sitewide for our app, and it will be available to every package that imports this package
type AppConfig struct {
	UseCache      bool
	TemplateCache map[string]*template.Template
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
	InProduction  bool
	Session       *scs.SessionManager
	MailChan      chan models.MailData
}
