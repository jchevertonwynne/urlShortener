package webserver

import (
    "github.com/dgrijalva/jwt-go"
    "html/template"
    "net/http"
    "time"
    "urlShortener/pkg/database"
)

type createUserInformation struct {
    ErrorHappened    bool
    Error            string
}

const createUserTemplateLocation = "pkg/webserver/templates/createUser.html"

var createUserTemplate = template.Must(template.ParseFiles(createUserTemplateLocation))

func showCreateUserPage(res http.ResponseWriter, req *http.Request) {
    info := new(createUserInformation)

    errorCookie, err := req.Cookie("error")
    if err == nil {
        info.ErrorHappened = true
        info.Error = errorCookie.Value
        http.SetCookie(res, &http.Cookie{
            Name:       "error",
            Value:      "",
            Expires:    time.Now(),
            Path:       routeMain,
        })
    }

    createUserTemplate.Execute(res, info)
}

func handleCreation(res http.ResponseWriter, req *http.Request) {
    req.ParseForm()

    usernames, ok := req.Form["username"]
    if !ok || len(usernames) == 0 || len(usernames[0]) == 0 {
        http.SetCookie(res, &http.Cookie{
            Name:       "error",
            Value:      "Username must be filled in",
            Expires:    time.Now().Add(time.Minute),
            Path:       routeMain,
        })
        http.Redirect(res, req, routeCreateUser, http.StatusSeeOther)
        return
    }

    passwords, ok := req.Form["password"]
    if !ok || len(passwords) == 0 || len(passwords[0]) == 0 {
        http.SetCookie(res, &http.Cookie{
            Name:       "error",
            Value:      "Password must be filled in",
            Expires:    time.Now().Add(time.Minute),
            Path:       routeMain,
        })
        http.Redirect(res, req, routeCreateUser, http.StatusSeeOther)
        return
    }

    username := usernames[0]
    password := passwords[0]

    err := database.AddUser(username, password)
    if err != nil {
        http.SetCookie(res, &http.Cookie{
            Name:       "error",
            Value:      err.Error(),
            Expires:    time.Now().Add(time.Minute),
            Path:       routeMain,
        })
        http.Redirect(res, req, routeCreateUser, http.StatusSeeOther)
        return
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "username": username,
    })
    signedString, err := token.SignedString(JWTSecret)
    if err != nil {
        http.SetCookie(res, &http.Cookie{
            Name:       "error",
            Value:      "err with signing cookie",
            Expires:    time.Now().Add(time.Minute),
            Path:       routeMain,
        })
        http.Redirect(res, req, routeLogin, http.StatusSeeOther)
        return
    }

    http.SetCookie(res, &http.Cookie{
        Name:       "login",
        Value:      signedString,
        Expires:    time.Now().Add(time.Hour),
        Path:       "/",
    })

    http.Redirect(res, req, routeMain, http.StatusSeeOther)
}