package main

import (
	"urlShortener/pkg/webserver"
)

func main() {
	s := webserver.Create()
	s.ListenAndServe()
}
