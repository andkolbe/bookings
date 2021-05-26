package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/andkolbe/bookings/internal/config"
	"github.com/andkolbe/bookings/internal/driver"
	"github.com/andkolbe/bookings/internal/forms"
	"github.com/andkolbe/bookings/internal/helpers"
	"github.com/andkolbe/bookings/internal/models"
	"github.com/andkolbe/bookings/internal/render"
	"github.com/andkolbe/bookings/internal/repository"
	"github.com/andkolbe/bookings/internal/repository/dbrepo"
)

// the repository used by the handlers
var Repo *Repository

// the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// creates a new repository
// when NewRepo is called, the app config and database connection pool are passed in
// the Repository type is populated with all of the info received as parameters and it handed back as a pointer to Repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestingRepo(a),
	}
}

// sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// home page handler
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "home.page.html", &models.TemplateData{})
}

// about page handler
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {

	// send the data  to the template
	render.Template(w, r, "about.page.html", &models.TemplateData{})
}

// reservation page handler
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// get the room id off of the session and check if it is valid
	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't find room")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.Room.RoomName = room.RoomName

	m.App.Session.Put(r.Context(), "reservation", res)

	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	// store in a data variable
	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r, "make-reservation.page.html", &models.TemplateData{
		// pass an empty form and the data variable to the template
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

// POST reservation page handler
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	// we start by pulling our reservation from the session
	// our reservation already has start date, end date, room id, and room name
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New("can't get from session"))
		return
	}

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse form!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// update reservation model
	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Phone = r.Form.Get("phone")
	reservation.Email = r.Form.Get("email")

	// create a new form
	form := forms.New(r.PostForm) // PostForm has all of the url values and their associated data

	// check if the incoming form has all of the fields filled out
	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	// if one of the fields on the form is not valid, repopulate the form with the data they entered and display the error message where it needs to be
	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		render.Template(w, r, "make-reservation.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	// save data to db and get a reservation back
	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't insert reservation into db")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't insert room restriction")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// send notifications - to guest
	htmlMessage := fmt.Sprintf(`
		<strong>Reservation Confirmation</strong><br>
		Dear %s, <br>
		This is to confirm your reservation from %s to %s.
	`, reservation.FirstName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))


	msg := models.MailData {
		To:      reservation.Email,
		From:    "bnbbooking@gmail.com",
		Subject: "Reservation Confirmation",
		Content: htmlMessage,
		Template: "basic.html",
	}
	// pass the msg into the app.MailChan channel
	m.App.MailChan <- msg


	// send notifications - to property owner
	htmlMessage = fmt.Sprintf(`
		<strong>Reservation Notification</strong><br>
		A reservation has been made for %s from %s to %s.
	`, reservation.Room.RoomName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))


	msg = models.MailData {
		To:      "bnbbooking@gmail.com",
		From:    "bnbbooking@gmail.com",
		Subject: "Reservation Notification",
		Content: htmlMessage,
	}
	// pass the msg into the app.MailChan channel
	m.App.MailChan <- msg

	m.App.Session.Put(r.Context(), "reservation", reservation)

	// redirect the user to a different page after submitting the form so they can't click the submut button twice
	// StatusSeeOther is response code 303
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// generals room page handler
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "generals.page.html", &models.TemplateData{})
}

// majors room page handler
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "majors.page.html", &models.TemplateData{})
}

// room availability page handler
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.html", &models.TemplateData{})
}

// POST room availability page handler
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start") // start and end match the names of the input fields on the search-availability.page.html
	end := r.Form.Get("end")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, start)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	endDate, err := time.Parse(layout, end)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	if len(rooms) == 0 {
		// neither room is available
		m.App.Session.Put(r.Context(), "error", "No availability")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	// create a data variable that is of string interface
	data := make(map[string]interface{})
	// store the rooms in that map
	data["rooms"] = rooms

	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	// store the start and end dates in the session
	m.App.Session.Put(r.Context(), "reservation", res)

	// pass the data to the template choose-room
	render.Template(w, r, "choose-room.page.html", &models.TemplateData{
		Data: data,
	})
}

// if you want to export a struct to JSON, the member names must start with a capital letter
type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// handles request for availability and sends JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	// need to parse request body to be able to write a test for it
	err := r.ParseForm()
	if err != nil {
		// can't parse form, so return appropriate JSON
		resp := jsonResponse{
			OK:      false,
			Message: "Internal server error",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	sd := r.Form.Get("start")
	ed := r.Form.Get("end")

	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	roomID, _ := strconv.Atoi(r.Form.Get("room_id"))

	available, _ := m.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomID)
	if err != nil {
		// can't parse form, so return appropriate JSON
		resp := jsonResponse{
			OK:      false,
			Message: "Error connecting to database",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	// create the struct type in a reusable variable
	resp := jsonResponse{
		OK:        available,
		Message:   "",
		StartDate: sd,
		EndDate:   ed,
		RoomID:    strconv.Itoa(roomID),
	}

	// marshall the resp into json
	out, _ := json.MarshalIndent(resp, "", "     ")
	// don't need an error check, since we handle all aspects of the JSON already

	// write the result as application/json to the web browser
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// contact page handler
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.html", &models.TemplateData{})
}

// displays the reservation summary page
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

	// format start and end dates from strings to time
	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	render.Template(w, r, "reservation-summary.page.html", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

// displays list of available rooms
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	// changed to this, so we can test it more easily
	// split the URL up by /, and grab the 3rd element
	exploded := strings.Split(r.RequestURI, "/")
	roomID, err := strconv.Atoi(exploded[2])
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "missing url parameter")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, err)
		return
	}

	res.RoomID = roomID

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// takes url parameters, builds session variable, and takes user to make res screen
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	// get the query parameters from the incoming request
	roomID, _ := strconv.Atoi(r.URL.Query().Get("id")) // need to convert id from string to an int
	sd := r.URL.Query().Get("s")
	ed := r.URL.Query().Get("e")

	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	var res models.Reservation

	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res.Room.RoomName = room.RoomName
	res.RoomID = roomID
	res.StartDate = startDate
	res.EndDate = endDate

	// store the data in the session
	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.html", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// handles logging the user in
func (m *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
	// good practice to renew the token every time a user logs in/out
	_ = m.App.Session.RenewToken(r.Context())

	// need to parse request body to be able to write a test for it
	err := r.ParseForm()
	if err != nil {
		// can't parse form, so return appropriate JSON
		// resp := jsonResponse{
		// 	OK:      false,
		// 	Message: "Internal server error",
		log.Println(err)
		}

		email := r.Form.Get("email")
		password := r.Form.Get("password")

		form := forms.New(r.PostForm)
		form.Required("email", "password")
		form.IsEmail("email")
		if !form.Valid() {
			render.Template(w, r, "login.page.html", &models.TemplateData{
				Form: form,
			})
			return
		}

		id, _, err := m.DB.Authenticate(email, password)
		if err != nil {
			log.Println(err)

			m.App.Session.Put(r.Context(), "error", "Invalid login credentials")
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		// if they are successfully authenticated, we store their id in the session
		m.App.Session.Put(r.Context(), "user_id", id)
		m.App.Session.Put(r.Context(), "flash", "Logged in successfully")
		http.Redirect(w, r, "/", http.StatusSeeOther)

		// out, _ := json.MarshalIndent(resp, "", "     ")
		// w.Header().Set("Content-Type", "application/json")
		// w.Write(out)
		// return
}

// logs a user out
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	// destroy the entire session data
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// logs a user out
func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	
	render.Template(w, r, "admin-dashboard.page.html", &models.TemplateData{})

}
