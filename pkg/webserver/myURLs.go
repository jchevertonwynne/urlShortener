package webserver

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"time"
	"urlShortener/pkg/database"
)

type myURLsInformation struct {
	DeletionHappened bool
	Deletion         string
	ErrorHappened    bool
	Error            string
	URLs             []database.Record
	LoggedInAs       string
}

const myURLsTemplateLocation = "pkg/webserver/templates/myURLs.html"

var myURLsTemplate = template.Must(template.ParseFiles(myURLsTemplateLocation))

func showMyLinksPage(res http.ResponseWriter, req *http.Request) {
	info := new(myURLsInformation)

	deletionCookie, err := req.Cookie("deletion")
	if err == nil {
		info.DeletionHappened = true
		info.Deletion = deletionCookie.Value
		http.SetCookie(res, &http.Cookie{
			Name:    "deletion",
			Value:   "",
			Expires: time.Now(),
			Path:    "/",
		})
	}

	errorCookie, err := req.Cookie("error")
	if err == nil {
		info.ErrorHappened = true
		info.Error = errorCookie.Value
		http.SetCookie(res, &http.Cookie{
			Name:    "error",
			Value:   "",
			Expires: time.Now(),
			Path:    routeMain,
		})
	}

	user, _ := verifyUsernameCookie(res, req)
	info.LoggedInAs = user
	urls, err := database.GetURLsOf(user)
	if err != nil {
		info.ErrorHappened = true
		info.Error = err.Error()
	}
	info.URLs = urls

	myURLsTemplate.Execute(res, info)
}

func deleteURLRouteHandler(res http.ResponseWriter, req *http.Request) {
	// TODO: user can delete their owned URLs
	vars := mux.Vars(req)
	shortened, _ := vars["key"]
	username, _ := verifyUsernameCookie(res, req)
	ok := database.VerifyOwns(username, shortened)
	if ok {
		err := database.DeleteURL(shortened)
		if err != nil {
			fmt.Println(err)
		}
		http.SetCookie(res, &http.Cookie{
			Name:    "deletion",
			Value:   "Deleted shortened URL",
			Expires: time.Now().Add(time.Minute),
			Path:    routeMain,
		})
	} else {
		http.SetCookie(res, &http.Cookie{
			Name:    "error",
			Value:   "URL not owned by you",
			Expires: time.Now().Add(time.Minute),
			Path:    routeMain,
		})
	}
	http.Redirect(res, req, routeMyLinks, http.StatusSeeOther)
}
