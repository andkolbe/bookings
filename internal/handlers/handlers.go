package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/andkolbe/bookings/internal/config"
	"github.com/andkolbe/bookings/internal/forms"
	"github.com/andkolbe/bookings/internal/helpers"
	"github.com/andkolbe/bookings/internal/models"
	"github.com/andkolbe/bookings/internal/render"
)

// the repository used by the handlers
var Repo *Repository

// the repository type
type Repository struct {
	App *config.AppConfig
}

// creates a new repository
func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

// sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// home page handler
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "home.page.html", &models.TemplateData{})
}

// about page handler
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {

	// send the data  to the template
	render.RenderTemplate(w, r, "about.page.html", &models.TemplateData{})
}

// reservation page handler
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	// create an empty reservation for the first time this page is displayed
	var emptyReservation models.Reservation
	// store in a data variable
	data := make(map[string]interface{})
	data["reservation"] = emptyReservation

	render.RenderTemplate(w, r, "make-reservation.page.html", &models.TemplateData{
		// pass an empty form and the data variable to the template
		Form: forms.New(nil),
		Data: data,
	})
}

// POST reservation page handler
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// store incoming reservation data that a user submitted in a variable
	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
	}

	// create a new form
	form := forms.New(r.PostForm) // PostForm has all of the url values and their associated data

	// check if the incoming form has all of the fields filled out
	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		render.RenderTemplate(w, r, "make-reservation.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	m.App.Session.Put(r.Context(), "reservation", reservation)

	// redirect the user to a different page after submitting the form so they can't click the submut button twice
	// StatusSeeOther is response code 303
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// generals room page handler
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "generals.page.html", &models.TemplateData{})
}

// majors room page handler
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "majors.page.html", &models.TemplateData{})
}

// room availability page handler
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "search-availability.page.html", &models.TemplateData{})
}

// POST room availability page handler
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start") // start and end match the names of the input fields on the search-availability.page.html
	end := r.Form.Get("end")

	w.Write([]byte(fmt.Sprintf("start date is %s and end date is %s", start, end)))
}

// if you want to export a struct to JSON, the member names must start with a capital letter
type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

// handles request for availability and sends JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	// create the struct type in a reusable variable
	resp := jsonResponse{
		OK:      true,
		Message: "Available!",
	}

	// marshall the resp into json
	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// write the result as application/json to the web browser
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// contact page handler
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "contact.page.html", &models.TemplateData{})
}

func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	// if it finds something called reservation in the session and it manages to assert it to type models.Reservation, ok will be true
	if !ok {
		m.App.ErrorLog.Println("Can't get error from session")
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect) // redirect them to the home page 
		return
	}

	m.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

 	render.RenderTemplate(w, r, "reservation-summary.page.html", &models.TemplateData{
		 Data: data,
	 })
}
