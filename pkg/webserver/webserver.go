package webserver

import (
    "github.com/gorilla/mux"
    "net/http"
    "time"
    "urlShortener/pkg/database"
)

const (
    routeMain = "/"
    routeLogin = "/login"
    routeCreateUser = "/createUser"
    routeRedirect = "/u/{key}"
)

var JWTSecret []byte

func Create() *http.Server {
    handler := mux.NewRouter()
    handler.HandleFunc(routeMain, homePageRouteHandler)
    handler.HandleFunc(routeLogin, loginRouteHandler)
    handler.HandleFunc(routeCreateUser, createUserHandler)
    handler.HandleFunc(routeRedirect, redirectRouteHandler)

    return &http.Server{
        Addr: ":8000",
        Handler: handler,
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

func redirectRouteHandler(res http.ResponseWriter, req *http.Request) {
    vars := mux.Vars(req)
    shortened, _ := vars["key"]
    url, err := database.GetUrl(shortened)
    if err != nil {
        http.SetCookie(res, &http.Cookie{
            Name:       "error",
            Value:      "shortened url not found",
            Expires:    time.Now().Add(time.Minute),
            Path:       routeMain,
        })
        http.Redirect(res, req, routeMain, http.StatusSeeOther)
        return
    }
    http.Redirect(res, req, url, http.StatusSeeOther)
}
