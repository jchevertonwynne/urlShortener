package webserver

import (
    "fmt"
    "github.com/dgrijalva/jwt-go"
    "html/template"
    "math/rand"
    "net/http"
    "net/url"
    "time"
    "urlShortener/pkg/database"
)

type homePageInformation struct {
    ErrorHappened    bool
    Error            string
    CreatedShortened bool
    Shortened        string
    LoggedIn         bool
    LoggedInAs       string
}

const homeTemplateLocation = "pkg/webserver/templates/home.html"

var homeTemplate = template.Must(template.ParseFiles(homeTemplateLocation))

func init() {
    rand.Seed(int64(time.Now().Nanosecond()))
}

func showHomePage(res http.ResponseWriter, req *http.Request) {
    info := new(homePageInformation)

    user, err := loggedIn(req)
    if err == nil {
        info.LoggedIn = true
        info.LoggedInAs = user
    }

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
    err := req.ParseForm()
    if err != nil {
        http.SetCookie(res, &http.Cookie{
            Name:       "error",
            Value:      err.Error(),
            Expires:    time.Now().Add(time.Minute),
            Path:       "/",
        })
        http.Redirect(res, req, "/", http.StatusSeeOther)
        return
    }

    userURLs, ok := req.Form["url"]
    if !ok || len(userURLs) == 0 || len(userURLs[0]) == 0 {
        http.SetCookie(res, &http.Cookie{
            Name:       "error",
            Value:      "no url sent in form",
            Expires:    time.Now().Add(time.Minute),
            Path:       "/",
        })
        http.Redirect(res, req, "/", http.StatusSeeOther)
        return
    }

    userURL := userURLs[0]

    _, err = url.ParseRequestURI(userURL)
    if err != nil {
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

    err = database.AddURL(userURL, shortened)
    if err != nil {
        http.SetCookie(res, &http.Cookie{
            Name:       "error",
            Value:      err.Error(),
            Expires:    time.Now().Add(time.Minute),
            Path:       "/",
        })
        return
    }
    http.SetCookie(res, &http.Cookie{
        Name:       "created",
        Value:      shortened,
        Expires:    time.Now().Add(time.Minute),
        Path:       "/",
    })

    user, err := loggedIn(req)
    if err == nil {
        database.Link(user, shortened)
    }

    http.Redirect(res, req, "/", http.StatusSeeOther)
}

func randomChars() string {
    chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
    res := make([]byte, 8)
    for i := 0; i < 8; i++ {
        res[i] = chars[rand.Intn(len(chars))]
    }
    return string(res)
}

func loggedIn(req *http.Request) (string, error) {
    loginCookie, err := req.Cookie("login")
    if err != nil {
        return "", err
    }
    val := loginCookie.Value
    token, err := jwt.Parse(val, func(token *jwt.Token) (interface{}, error) {
        _, ok := token.Method.(*jwt.SigningMethodHMAC)
        if !ok {
            return nil, fmt.Errorf("method not valid")
        }
        return JWTSecret, nil
    })
    if err != nil {
        return "", err
    }
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || !token.Valid {
        return "", fmt.Errorf("user could not be verified")
    }
    return claims["username"].(string), nil
}