package webserver

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"time"
	"urlShortener/pkg/database"
)

const (
	routeMain       = "/"
	routeLogin      = "/login"
	routeLogout     = "/logout"
	routeCreateUser = "/createUser"
	routeDeleteUser = "/deleteUser"
	routeMyLinks    = "/profile"
	routeRedirect   = "/u/{key}"
	routeDeleteURL  = "/d/{key}"
)

var jwtSecret []byte

func Run(secret string) {
	jwtSecret = []byte(secret)
	server := create()
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}

func create() *http.Server {
	handler := mux.NewRouter()
	handler.HandleFunc(routeMain, homePageRouteHandler)
	handler.HandleFunc(routeLogin, mustBeLoggedOut(loginRouteHandler))
	handler.HandleFunc(routeLogout, logoutRouteHandler)
	handler.HandleFunc(routeCreateUser, mustBeLoggedOut(createUserHandler))
	handler.HandleFunc(routeDeleteUser, mustBeLoggedIn(deleteUserHandler))
	handler.HandleFunc(routeMyLinks, mustBeLoggedIn(myLinksHandler))
	handler.HandleFunc(routeRedirect, redirectRouteHandler)
	handler.HandleFunc(routeDeleteURL, deleteURLRouteHandler)

	return &http.Server{
		Addr:    "0.0.0.0:8000",
		Handler: handler,
	}
}

func mustBeLoggedIn(f http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		_, err := verifyUsernameCookie(res, req)
		if err != nil {
			http.SetCookie(res, &http.Cookie{
				Name:    "error",
				Value:   fmt.Sprintf("You must be logged in to access %s", req.URL.Path),
				Expires: time.Now().Add(time.Minute),
				Path:    routeMain,
			})
			http.Redirect(res, req, routeMain, http.StatusSeeOther)
			return
		}
		f(res, req)
	}
}

func mustBeLoggedOut(f http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		_, err := verifyUsernameCookie(res, req)
		if err == nil {
			http.SetCookie(res, &http.Cookie{
				Name:    "error",
				Value:   fmt.Sprintf("You must be logged out to access %s", req.URL.Path),
				Expires: time.Now().Add(time.Minute),
				Path:    routeMain,
			})
			http.Redirect(res, req, routeMain, http.StatusSeeOther)
			return
		}
		f(res, req)
	}
}

func homePageRouteHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		showHomePage(res, req)
	case http.MethodPost:
		handleURLUpload(res, req)
	default:
		http.Redirect(res, req, routeMain, http.StatusSeeOther)
	}
}

func loginRouteHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		showLoginPage(res, req)
	case http.MethodPost:
		handleLogin(res, req)
	default:
		http.Redirect(res, req, routeMain, http.StatusSeeOther)
	}
}

func logoutRouteHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		handleLogout(res, req)
	default:
		http.Redirect(res, req, routeMain, http.StatusSeeOther)
	}
}

func createUserHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		showCreateUserPage(res, req)
	case http.MethodPost:
		handleCreation(res, req)
	default:
		http.Redirect(res, req, routeMain, http.StatusSeeOther)
	}
}

func deleteUserHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		handleDeleteUser(res, req)
	default:
		http.Redirect(res, req, routeMain, http.StatusSeeOther)
	}
}

func myLinksHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		showMyLinksPage(res, req)
	default:
		http.Redirect(res, req, routeMain, http.StatusSeeOther)
	}
}

func redirectRouteHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	shortened, _ := vars["key"]
	url, err := database.GetUrl(shortened)
	if err != nil {
		http.SetCookie(res, &http.Cookie{
			Name:    "error",
			Value:   "shortened url not found",
			Expires: time.Now().Add(time.Minute),
			Path:    routeMain,
		})
		http.Redirect(res, req, routeMain, http.StatusSeeOther)
		return
	}
	http.Redirect(res, req, url.Long, http.StatusSeeOther)
}

func deleteURLRouteHandler(res http.ResponseWriter, req *http.Request) {
	// TODO: user can delete their owned URLs
}
