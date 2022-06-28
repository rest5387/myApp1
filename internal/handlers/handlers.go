package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/rest5387/myApp1/internal/helpers"

	"github.com/rest5387/myApp1/internal/forms"

	"github.com/rest5387/myApp1/internal/models"

	"github.com/rest5387/myApp1/internal/config"

	"github.com/rest5387/myApp1/internal/render"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
}

// NewRepo create a new repository
func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

//NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

//Home is the home page handler
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	// remoteIP := r.RemoteAddr
	// m.App.Session.Put(r.Context(), "remote_ip", remoteIP)
	// fmt.Println("Home put remote addr", m.App.Session.GetString(r.Context(), "remote_ip"))
	// fmt.Println("remote_ip", r.RemoteAddr)
	// fmt.Println("(Home func)Session key remote_ip exists?", m.App.Session.Exists(r.Context(), "remote_ip"))

	render.RenderTemplate(w, r, "home.page.tmpl", &models.TemplateData{})
}

//About is the about page handler
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	//perform some logic
	// stringMap := make(map[string]string)
	// stringMap["test"] = "Hello, again."
	// fmt.Println("Session lifetime", m.App.Session.Lifetime)
	// fmt.Println("Session cookie name", m.App.Session.Cookie.Name)
	// fmt.Println("Session key remote_ip exists?", m.App.Session.Exists(r.Context(), "remote_ip"))
	// remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")
	// stringMap["remote_ip"] = remoteIP

	// fmt.Println("Actual remote addr", r.RemoteAddr)
	// fmt.Println("About get remote addr", remoteIP, "\nstringmap ", stringMap["remote_ip"])

	// send the data to the template
	render.RenderTemplate(w, r, "about.page.tmpl", &models.TemplateData{
		// StringMap: stringMap,
	})
}

// Reservation renders the make-reservation page and displays form
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	var emptyReservation models.Reservation
	data := make(map[string]interface{})
	data["reservation"] = emptyReservation

	// send the data to the template
	render.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

// PostReservation handles the posting of a reservation form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	// call r.ParseForm at first to make r.Form/r.PostForm be available
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
	}

	// Use self-defined Form structure to wrap request's Form/PostForm (url.Values)
	form := forms.New(r.PostForm)

	// form.Has("first_name", r)
	form.Required("first_name", "last_name", "email")
	form.MiniLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		render.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	// put reservation data to the session
	m.App.Session.Put(r.Context(), "reservation", reservation)

	//redirect user after Post request
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// Generals renders the room page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	// send the data to the template
	render.RenderTemplate(w, r, "generals.page.tmpl", &models.TemplateData{})
}

// Majors renders the room page
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	// send the data to the template
	render.RenderTemplate(w, r, "majors.page.tmpl", &models.TemplateData{})
}

// Availability renders the search availability page
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	// send the data to the template
	render.RenderTemplate(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

// PostAvailability renders the search availability page
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	// send the data to the template
	start := r.Form.Get("start")
	end := r.Form.Get("end")
	w.Write([]byte(fmt.Sprintf("Start date is %s, End date is %s", start, end)))
}

type jsonResponse struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

// AvailabilityJSON handles request for availability and send JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	resp := jsonResponse{
		Ok:      true,
		Message: "Available!",
	}

	out, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	log.Println(string(out))

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// Contact renders the search availability page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	// send the data to the template
	render.RenderTemplate(w, r, "contact.page.tmpl", &models.TemplateData{})
}

// ReservationSummary renders the search availability page
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.ErrorLog.Println("can't get error from session")
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Remove(r.Context(), "reservation")
	data := make(map[string]interface{})
	data["reservation"] = reservation

	// send the data to the template
	render.RenderTemplate(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data: data,
	})
}
