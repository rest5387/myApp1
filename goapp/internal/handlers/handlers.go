package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/justinas/nosurf"
	"github.com/rest5387/myApp1/goapp/internal/helpers"
	"github.com/rest5387/myApp1/goapp/internal/repository"
	"github.com/rest5387/myApp1/goapp/internal/repository/dbrepo"
	"golang.org/x/crypto/bcrypt"

	"github.com/rest5387/myApp1/goapp/internal/forms"

	"github.com/rest5387/myApp1/goapp/internal/models"

	"github.com/rest5387/myApp1/goapp/internal/config"

	"github.com/rest5387/myApp1/goapp/internal/driver"
	"github.com/rest5387/myApp1/goapp/internal/render"
)

// Repo the repository used by the handlers
var Repo *Repository
var cardIdxMutex sync.Mutex

const getCardsAtOnce int = 3

// Repository is the repository type
type Repository struct {
	App        *config.AppConfig
	SQLDB      repository.SQLDatabaseRepo
	Neo4j      repository.Neo4jRepo
	RedisCache repository.RedisRepo
}

// NewRepo create a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App:        a,
		SQLDB:      dbrepo.NewPostgresRepo(db.SQL, a),
		Neo4j:      dbrepo.NewNeo4jRepo(db.Neo4j, a),
		RedisCache: dbrepo.NewRedisRepo(db.RedisCache, a),
	}
}

//NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

//Home is the home page handler
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {

	// if not log in, render login page
	if !m.App.Session.Exists(r.Context(), "uid") {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: forms.New(nil),
		})
		return
	}

	uid, ok := m.App.Session.Get(r.Context(), "uid").(int)
	if !ok {
		m.App.ErrorLog.Println("can't get error from session")
		m.App.Session.Put(r.Context(), "error", "Can't get uid from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// send user's PIDs to front-end
	// pids, err := m.SQLDB.SearchPIDsByUID(uid)
	// if err != nil {
	// 	m.App.ErrorLog.Println("can't get PIDs from DB")
	// 	m.App.Session.Put(r.Context(), "error", "Can't get posts list")
	// 	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	// 	return
	// }

	// Get all follows UID
	follows, err := m.Neo4j.GetAllFollowedUID(uid)
	if err != nil {
		http.Error(w, "get all followed uids error", http.StatusInternalServerError)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Get PIDs follows wrote in recent days
	followsPids, err := m.SQLDB.GetFollowsPIDS(follows)
	if err != nil {
		http.Error(w, "get all follows pids error", http.StatusInternalServerError)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Put(r.Context(), "card-list", followsPids)
	render.Template(w, r, "home.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})

}

// PostHome is the home page new post handler
func (m *Repository) PostHome(w http.ResponseWriter, r *http.Request) {
	if !m.App.Session.Exists(r.Context(), "uid") {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: forms.New(nil),
		})
		return
	}

	uid, ok := m.App.Session.Get(r.Context(), "uid").(int)
	if !ok {
		m.App.ErrorLog.Println("can't get error from session")
		m.App.Session.Put(r.Context(), "error", "Can't get uid from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	form := forms.New(r.PostForm)

	form.Required("textarea1")

	if !form.Valid() {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	text := form.Values.Get("textarea1")

	post := models.Post{
		UID:     uid,
		Likes:   0,
		Content: text,
	}
	_, err = m.SQLDB.InsertPost(post)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "System error! Try post again later!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

// User is the user page handler which render the personal page and
// pass visitedUID to front-end.
func (m *Repository) User(w http.ResponseWriter, r *http.Request) {

	// if not log in, render login page
	if !m.App.Session.Exists(r.Context(), "uid") {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: forms.New(nil),
		})
		return
	}

	uid, ok := m.App.Session.Get(r.Context(), "uid").(int)
	if !ok {
		m.App.ErrorLog.Println("can't get error from session")
		m.App.Session.Put(r.Context(), "error", "Can't get uid from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	visitedUid := r.Context().Value("userid").(int)

	m.SQLDB.SearchUserByUID(visitedUid)

	pids, err := m.SQLDB.SearchPIDsByUID(visitedUid)
	if err != nil {
		m.App.ErrorLog.Println("can't get PIDs from DB")
		m.App.Session.Put(r.Context(), "error", "Can't get posts list")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	m.App.Session.Put(r.Context(), "card-list", pids)

	data := make(map[string]interface{})
	data["UserID"] = visitedUid
	data["SelfVisit"] = false
	if uid == visitedUid {
		data["SelfVisit"] = true
	}

	render.Template(w, r, "personal.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})

}

type jsonPostResponse struct {
	PID         int       `json:"pid"`
	UID         int       `json:"uid"`
	Likes       int       `json:"likes"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	Content     string    `json:"content"`
	Created_at  time.Time `json:"created_at"`
	Updated_at  time.Time `json:"updated_at"`
	Error       error     `json:"error"`
	ResponseEnd bool      `json:"end"`
	Edit        bool      `json:"edit"`
}

// PostAJAX handles AJAX post card request and send a JSON response if the request
// and DB operations are all success.
func (m *Repository) PostCardAJAX(w http.ResponseWriter, r *http.Request) {
	if !m.App.Session.Exists(r.Context(), "uid") {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: forms.New(nil),
		})
		return
	}

	uid, ok := m.App.Session.Get(r.Context(), "uid").(int)
	if !ok {
		m.App.ErrorLog.Println("can't get error from session")
		m.App.Session.Put(r.Context(), "error", "Can't get uid from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	form := forms.New(r.PostForm)

	form.Required("textarea1")

	if !form.Valid() {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	text := form.Values.Get("textarea1")

	post := models.Post{
		UID:     uid,
		Likes:   0,
		Content: text,
	}

	pid, err := m.SQLDB.InsertPost(post)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "System error! Try post again later!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	newpost, err := m.SQLDB.SearchPostByPID(pid)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	user, err := m.SQLDB.SearchUserByUID(uid)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	resp := jsonPostResponse{
		PID:        newpost.ID,
		UID:        newpost.UID,
		Likes:      newpost.Likes,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Content:    newpost.Content,
		Created_at: newpost.Created_at,
		Updated_at: newpost.Updated_at,
		Edit:       true,
	}
	writeJsonResponse(w, resp)
}

// LoadPostAJAX handles GET cards request and return a JSON style response
// if the request and DB operations are success.
func (m *Repository) GetCardAJAX(w http.ResponseWriter, r *http.Request) {

	resps := make([]jsonPostResponse, 0, getCardsAtOnce)

	if !m.App.Session.Exists(r.Context(), "uid") {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: forms.New(nil),
		})
		return
	}
	// get uid for check there is an user logged in
	uid, ok := m.App.Session.Get(r.Context(), "uid").(int)
	if !ok {
		m.App.ErrorLog.Println("can't get error from session")
		m.App.Session.Put(r.Context(), "error", "Can't get uid from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if !m.App.Session.Exists(r.Context(), "card-list") {
		m.App.ErrorLog.Println("can't get error from session")
		m.App.Session.Put(r.Context(), "error", "Can't get card-list from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	offset := r.Context().Value("offset").(int)
	visitedUID := r.Context().Value("userid").(int)
	card_list := m.App.Session.Get(r.Context(), "card-list").([]int)
	var pidsErr error
	if visitedUID != 0 {
		card_list, pidsErr = m.SQLDB.SearchPIDsByUID(visitedUID)
		if pidsErr != nil {
			m.App.ErrorLog.Println("can't get PIDs from DB")
			m.App.Session.Put(r.Context(), "error", "Can't get posts list")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
	}

	if offset >= len(card_list) {
		return
	}

	var resp jsonPostResponse
	for idx := 0; idx < getCardsAtOnce && ((offset + idx) < len(card_list)); idx++ {
		key := strconv.Itoa(card_list[offset+idx])
		if m.RedisCache.Exists(key) {
			if err := m.RedisCache.Get(key, &resp); err != nil {
				return
			}
			fmt.Println("Cache hit! PID: ", resp.PID)
		} else {

			post, err := m.SQLDB.SearchPostByPID(card_list[offset+idx])
			if err != nil {
				m.App.ErrorLog.Println("can't get Post from DB")
				m.App.Session.Put(r.Context(), "error", "Can't get post content")
				http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
				return
			}
			user, err := m.SQLDB.SearchUserByUID(post.UID)
			if err != nil {
				m.App.ErrorLog.Println("can't get user-info from DB")
				m.App.Session.Put(r.Context(), "error", "Can't get user")
				http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
				return
			}
			resp = jsonPostResponse{
				PID:         card_list[offset+idx],
				UID:         post.UID,
				Likes:       post.Likes,
				FirstName:   user.FirstName,
				LastName:    user.LastName,
				Content:     post.Content,
				Created_at:  post.Created_at,
				Updated_at:  post.Updated_at,
				ResponseEnd: false,
				Edit:        false,
			}

			fmt.Println("Cache miss! PID: ", resp.PID)
		}
		if (offset + idx) == (len(card_list) - 1) {
			resp.ResponseEnd = true
		}

		if resp.UID == uid {
			resp.Edit = true
		}
		resps = append(resps, resp)
		m.RedisCache.Set(key, &resp)

	}

	writeJsonResponse(w, resps)
}

type jsonPutBody struct {
	Content   string `json:"content"`
	CsrfToken string `json:"csrf_token"`
}
type jsonPutResponse struct {
	IsUpdated bool   `json:"isUpdated"`
	Content   string `json:"content"`
}
type jsonDeleteBody struct {
	CsrfToken string `json:"csrf_token"`
}
type jsonDeleteResponse struct {
	IsDeleted bool   `json:"isDeleted"`
	Content   string `json:"content"`
}
type jsonUserResponse struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Followed  bool   `json:"followed"`
}
type jsonFollowBody struct {
	CsrfToken string `json:"csrf_token"`
}
type jsonFollowResponse struct {
	Success bool `json:"success"`
}

// PutAJAX handles AJAX Update card request and send a JSON response if the request
// and DB operations are all success.
func (m *Repository) PutCardAJAX(w http.ResponseWriter, r *http.Request) {

	var data jsonPutBody
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read request bidy error", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(requestBody, &data)
	if err != nil {
		http.Error(w, "parse json body error", http.StatusInternalServerError)
		return
	}

	if !nosurf.VerifyToken(nosurf.Token(r), data.CsrfToken) {
		http.Error(w, "CSRF token incorrect", http.StatusBadRequest)
		return
	}

	if !m.App.Session.Exists(r.Context(), "uid") {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	uid, ok := m.App.Session.Get(r.Context(), "uid").(int)
	if !ok {
		m.App.ErrorLog.Println("can't get error from session")
		m.App.Session.Put(r.Context(), "error", "Can't get uid from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	pid := r.Context().Value("pid").(int)
	post, _ := m.SQLDB.SearchPostByPID(pid)
	if post.UID != uid {
		http.Error(w, "user access level incorrect", http.StatusBadRequest)
		return
	}
	post.Content = data.Content

	resp := jsonPutResponse{
		IsUpdated: true,
		Content:   data.Content,
	}
	err = m.SQLDB.UpdatePostByPID(pid, *post)
	if err != nil {
		resp.IsUpdated = false
		resp.Content = ""
	}

	writeJsonResponse(w, resp)
}

// DeleteAJAX handles AJAX Delete card request and send a JSON response if the request
// and DB operations are all success.
func (m *Repository) DeleteCardAJAX(w http.ResponseWriter, r *http.Request) {

	var data jsonDeleteBody
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read request bidy error", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(requestBody, &data)
	if err != nil {
		http.Error(w, "parse json body error", http.StatusInternalServerError)
		return
	}

	if !nosurf.VerifyToken(nosurf.Token(r), data.CsrfToken) {
		http.Error(w, "CSRF token incorrect", http.StatusBadRequest)
		return
	}

	if !m.App.Session.Exists(r.Context(), "uid") {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	uid, ok := m.App.Session.Get(r.Context(), "uid").(int)
	if !ok {
		m.App.ErrorLog.Println("can't get error from session")
		m.App.Session.Put(r.Context(), "error", "Can't get uid from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	pid := r.Context().Value("pid").(int)
	post, _ := m.SQLDB.SearchPostByPID(pid)
	if post.UID != uid {
		http.Error(w, "user access level incorrect", http.StatusBadRequest)
		return
	}

	resp := jsonDeleteResponse{
		IsDeleted: true,
		Content:   "",
	}
	err = m.SQLDB.DeletePostByPID(pid)
	if err != nil {
		fmt.Println(err.Error())
		resp.IsDeleted = false
		resp.Content = ""
	}

	writeJsonResponse(w, resp)
}

func (m *Repository) GetUser(w http.ResponseWriter, r *http.Request) {
	if !m.App.Session.Exists(r.Context(), "uid") {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: forms.New(nil),
		})
		return
	}
	uid := m.App.Session.Get(r.Context(), "uid").(int)
	visitedUid := r.Context().Value("userid").(int)
	visitedUser, err := m.SQLDB.SearchUserByUID(visitedUid)
	if err != nil {
		m.App.ErrorLog.Println("can't get user-info from SQL DB")
		m.App.Session.Put(r.Context(), "error", "Can't get user")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	followed, err := m.Neo4j.SearchFollow(uid, visitedUid)
	if err != nil {
		m.App.ErrorLog.Println("can't get user-follow-ship from Neo4j DB")
		m.App.Session.Put(r.Context(), "error", "Can't get user")
		url := fmt.Sprintf("/userid=%d", visitedUid)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		return
	}
	resp := jsonUserResponse{
		FirstName: visitedUser.FirstName,
		LastName:  visitedUser.LastName,
		Followed:  followed,
	}
	writeJsonResponse(w, resp)
}

func (m *Repository) PostFollow(w http.ResponseWriter, r *http.Request) {
	var data jsonFollowBody
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read request body error", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(requestBody, &data)
	if err != nil {
		http.Error(w, "parse json body error", http.StatusInternalServerError)
		return
	}

	if !nosurf.VerifyToken(nosurf.Token(r), data.CsrfToken) {
		http.Error(w, "CSRF token incorrect", http.StatusBadRequest)
		return
	}

	if !m.App.Session.Exists(r.Context(), "uid") {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: forms.New(nil),
		})
		return
	}
	uid := m.App.Session.Get(r.Context(), "uid").(int)
	visitedUid := r.Context().Value("userid").(int)

	resp := jsonFollowResponse{
		Success: true,
	}
	err = m.Neo4j.InsertFollow(uid, visitedUid)
	if err != nil {
		resp.Success = false
	}
	writeJsonResponse(w, resp)
}

func (m *Repository) DeleteFollow(w http.ResponseWriter, r *http.Request) {
	var data jsonFollowBody
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read request body error", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(requestBody, &data)
	if err != nil {
		http.Error(w, "parse json body error", http.StatusInternalServerError)
		return
	}

	if !nosurf.VerifyToken(nosurf.Token(r), data.CsrfToken) {
		http.Error(w, "CSRF token incorrect", http.StatusBadRequest)
		return
	}

	if !m.App.Session.Exists(r.Context(), "uid") {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: forms.New(nil),
		})
		return
	}
	uid := m.App.Session.Get(r.Context(), "uid").(int)
	visitedUid := r.Context().Value("userid").(int)

	resp := jsonFollowResponse{
		Success: true,
	}
	err = m.Neo4j.DeleteFollow(uid, visitedUid)
	if err != nil {
		resp.Success = false
	}

	writeJsonResponse(w, resp)
}

//Login is the login page handler
func (m *Repository) Login(w http.ResponseWriter, r *http.Request) {
	if !m.App.Session.Exists(r.Context(), "uid") {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: forms.New(nil),
		})
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (m *Repository) PostLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	form := forms.New(r.PostForm)

	form.Required("email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	// Select password hash with email from users table.
	// Compare password and hash to certify the login user's correctness.
	// If result is correct, login success and render home post-wall page,
	// otherwise back to login page and show the reason to user.

	user, err := m.SQLDB.SearchUserByEmail(email)
	if err != nil {
		//database error: user with Email not found
		//redirect to login page
		fmt.Printf("Render login page again to user\n")
		m.App.Session.Put(r.Context(), "error", fmt.Sprintf("The user with \"%s\" is not found. Try again or sign up a new account!", email))
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: forms.New(nil),
			// Error: fmt.Sprintf("The user with %s is not found.Try again or sign up a new account!", email),
		})
		return
	}

	// fmt.Printf("password hash : %s\n", user.PasswordHash)
	// if bcrypt.CompareHashAndPassword()
	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
	if err != nil {
		// password is incorrect
		// render
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	// login success
	// store user_id into session and show home page(post wall) or personal page.
	// redirect to home page and show the post wall.
	m.App.Session.Put(r.Context(), "uid", user.ID)

	http.Redirect(w, r, "/", http.StatusSeeOther)

	// w.Write([]byte(fmt.Sprintf("Post Login\n ---- email: %s\n ---- password: %s", email, password)))
	// render.RenderTemplate(w, r, "login.page.tmpl", &models.TemplateData{})
}

//Logout is the log-out operation handler
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	// if not log in, render login page
	if m.App.Session.Exists(r.Context(), "uid") {
		m.App.Session.Remove(r.Context(), "uid")
	}

	// render.RenderTemplate(w, r, "login.page.tmpl", &models.TemplateData{
	// 	Form: forms.New(nil),
	// })
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (m *Repository) SignUp(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "signup.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

func (m *Repository) PostSignUp(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	form := forms.New(r.PostForm)

	form.Required("email", "password1", "password2", "first_name", "last_name")
	form.IsEmail("email")
	form.InputDoubleCheck("password1", "password2")

	if !form.Valid() {
		render.Template(w, r, "signup.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}
	// fmt.Printf("Form input has no error.\n")

	email := r.Form.Get("email")

	_, err = m.SQLDB.SearchUserByEmail(email)

	if err == nil {
		fmt.Printf("This email has been registerd an account.\n")
		m.App.Session.Put(r.Context(), "error", "This email has been registerd. Log in or try another one.")
		render.Template(w, r, "signup.page.tmpl", &models.TemplateData{
			Form: forms.New(nil),
		})
		return
	}

	// Insert a new user to DB users table.
	password := r.Form.Get("password1")
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "System error! Try again later!")
		render.Template(w, r, "signup.page.tmpl", &models.TemplateData{
			Form: forms.New(nil),
		})
		return
	}
	newUser := models.User{
		FirstName:    r.Form.Get("first_name"),
		LastName:     r.Form.Get("last_name"),
		Email:        email,
		PasswordHash: hash,
	}

	err = m.SQLDB.InsertUser(newUser)
	if err != nil {
		// m.App.Session.Put(r.Context(), "error", err.Error())
		m.App.Session.Put(r.Context(), "error", "System error! Try Sign up again later!")
		render.Template(w, r, "signup.page.tmpl", &models.TemplateData{
			Form: forms.New(nil),
		})
		return
	}

	// redirect to home page and show the post wall.
	user, err := m.SQLDB.SearchUserByEmail(email)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "System error! Sign up Success but Try log in later!")
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: forms.New(nil),
		})
		return
	}
	// insert user into neo4j DB for follow function
	err = m.Neo4j.InsertUser(newUser)
	if err != nil {
		// m.App.Session.Put(r.Context(), "error", err.Error())
		m.App.Session.Put(r.Context(), "error", "System error! Try Sign up again later!")
		render.Template(w, r, "signup.page.tmpl", &models.TemplateData{
			Form: forms.New(nil),
		})
		return
	}
	m.App.Session.Put(r.Context(), "uid", user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func writeJsonResponse(w http.ResponseWriter, resp any) {
	out, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Access-Controll-Allow-Origins", "*")
	w.Write(out)
}
