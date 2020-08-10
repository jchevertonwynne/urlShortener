package main

import (
	"flag"
	"fmt"
	"urlShortener/pkg/database"
	"urlShortener/pkg/webserver"
)

func main() {
	username := flag.String("username", "neo4j", "username for neo4j instance")
	password := flag.String("password", "neo4j", "password for neo4j instance")
	secret := flag.String("secret", "potato", "secret for jwt signing")
	flag.Parse()
	err := database.Init(*username, *password)
	if err != nil {
		fmt.Println(err)
		return
	}
	webserver.Run(*secret)
}
