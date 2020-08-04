package webserver

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	homeTemplateLocation = "pkg/webserver/templates/home.html"
	loginTemplateLocation = "pkg/webserver/templates/home.html"
)

func init() {
	rand.Seed(int64(time.Now().Nanosecond()))
}

var (
	homeTemplate = template.Must(template.ParseFiles(homeTemplateLocation))
	loginTemplate = template.Must(template.ParseFiles(loginTemplateLocation))
	urls         = make(map[string]string)
	mu           = new(sync.Mutex)
)

type homePageInformation struct {
	ErrorHappened    bool
	Error            string
	CreatedShortened bool
	Shortened        string
}

func Create() *http.Server {
	handler := mux.NewRouter()
	handler.HandleFunc("/", mainRouteHandler)
	handler.HandleFunc("/login", loginHandler)
	handler.HandleFunc("/u/{key}", handleRedirect)

	return &http.Server{
		Addr: ":8000",
		Handler: handler,
	}
}

func loginHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		showLoginPage(res, req)
	case http.MethodPost:
		handleLogin(res, req)
	default:
		http.Redirect(res, req, "/", http.StatusSeeOther)
	}
}

func handleLogin(res http.ResponseWriter, req *http.Request) {
	loginTemplate.Execute(res, "lol")
}

func showLoginPage(res http.ResponseWriter, req *http.Request) {

}

func handleRedirect(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	shortened, _ := vars["key"]
	mu.Lock()
	url, ok := urls[shortened]
	mu.Unlock()
	if !ok {
		fmt.Println("setting cookie")
		http.SetCookie(res, &http.Cookie{
			Name:       "error",
			Value:      "shortened url not found",
			Expires:    time.Now().Add(time.Minute),
			Path:       "/",
		})
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	http.Redirect(res, req, url, http.StatusSeeOther)
}

func mainRouteHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		showHomePage(res, req)
	case http.MethodPost:
		handleURLUpload(res, req)
	default:
		http.Redirect(res, req, "/", http.StatusSeeOther)
	}
}

func showHomePage(res http.ResponseWriter, req *http.Request) {
	info := homePageInformation{}

	errorCookie, err := req.Cookie("error")
	if err == nil {
		info.ErrorHappened = true
		info.Error = errorCookie.Value
		http.SetCookie(res, &http.Cookie{
			Name:       "error",
			Value:      "",
			Expires:    time.Now(),
			Path:       "/",
		})
	}

	creationCookie, err := req.Cookie("created")
	if err == nil {
		info.CreatedShortened = true
		info.Shortened = creationCookie.Value
		http.SetCookie(res, &http.Cookie{
			Name:       "created",
			Value:      "",
			Expires:    time.Now(),
			Path:       "/",
		})
	}

	homeTemplate.Execute(res, info)
}

func handleURLUpload(res http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	userURL, ok := req.Form["url"]
	if !ok || len(userURL) == 0 {
		fmt.Println("no userURL sent in form")
		http.SetCookie(res, &http.Cookie{
			Name:       "error",
			Value:      "no url sent in form",
			Expires:    time.Now().Add(time.Minute),
			Path:       "/",
		})
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	_, err := url.ParseRequestURI(userURL[0])
	if err != nil {
		fmt.Println("invalid url provided")
		http.SetCookie(res, &http.Cookie{
			Name:       "error",
			Value:      "please enter a valid url",
			Expires:    time.Now().Add(time.Minute),
			Path:       "/",
		})
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	shortened := randomChars()
	mu.Lock()
	urls[shortened] = userURL[0]
	mu.Unlock()
	http.SetCookie(res, &http.Cookie{
		Name:       "created",
		Value:      shortened,
		Expires:    time.Now().Add(time.Minute),
		Path:       "/",
	})
	http.Redirect(res, req, "/", http.StatusSeeOther)
}

func randomChars() string {
	chars := "abcdefghijklmnopqrstuvwxyz"
	res := make([]byte, 8)
	for i := 0; i < 8; i++ {
		res[i] = chars[rand.Intn(len(chars))]
	}
	return string(res)
}