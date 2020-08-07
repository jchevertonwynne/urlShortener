package webserver

import (
    "html/template"
    "net/http"
    "urlShortener/pkg/database"
)

type myURLsInformation struct {
    ErrorHappened bool
    Error         string
    URLs          []database.Record
    LoggedInAs    string
}

const myURLsTemplateLocation = "pkg/webserver/templates/myURLs.html"

var myURLsTemplate = template.Must(template.ParseFiles(myURLsTemplateLocation))

func showMyLinksPage(res http.ResponseWriter, req *http.Request) {
    info := new(myURLsInformation)

    user, _ := loggedIn(req)
    info.LoggedInAs = user
    urls, err := database.GetURLsOf(user)
    if err != nil {
        info.ErrorHappened = true
        info.Error = err.Error()
    }
    info.URLs = urls

    myURLsTemplate.Execute(res, info)
}