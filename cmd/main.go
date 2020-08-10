package main

import (
	"flag"
	"urlShortener/pkg/database"
	"urlShortener/pkg/webserver"
)

func main() {
	username := flag.String("username", "neo4j", "username for neo4j instance")
	password := flag.String("password", "neo4j", "password for neo4j instance")
	secret := flag.String("secret", "potato", "secret for jwt signing")
	flag.Parse()
	database.Init(*username, *password)
	webserver.Run(*secret)
}
