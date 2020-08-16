package database

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Record struct {
	Short   string
	Long    string
	Created time.Time
}

type User struct {
	Username string
	Password string
	Created  time.Time
}

const (
	serverURL    = "bolt://localhost:7687"
	databaseName = "neo4j"
	bcryptCost   = 10
)

var driver neo4j.Driver

func Init(username, password string) error {
	d, err := neo4j.NewDriver(serverURL, neo4j.BasicAuth(username, password, ""), func(config *neo4j.Config) {
		config.Encrypted = false
	})
	if err != nil {
		return err
	}
	driver = d
	return nil
}

func AddURL(long, short string) error {
	sessionConfig := neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: databaseName,
	}
	session, err := driver.NewSession(sessionConfig)
	if err != nil {
		return err
	}
	defer session.Close()

	data := map[string]interface{}{"long": long, "short": short}
	res, err := session.Run("CREATE (u:URL {long:$long, short:$short, created: datetime({ timezone: 'Europe/London' })})", data)
	if err != nil {
		return err
	}

	return res.Err()
}

func GetUrl(short string) (Record, error) {
	sessionConfig := neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: databaseName,
	}
	session, err := driver.NewSession(sessionConfig)
	if err != nil {
		return Record{}, err
	}
	defer session.Close()

	data := map[string]interface{}{"short": short}
	res, err := session.Run("MATCH (u:URL {short:$short}) RETURN u LIMIT 1", data)
	if err != nil {
		return Record{}, err
	}

	for res.Next() {
		node := res.Record().GetByIndex(0).(neo4j.Node)
		record, err := ParseRecord(node)
		if err != nil {
			continue
		}
		return record, nil
	}

	return Record{}, fmt.Errorf("url not found")
}

func GetUser(username string) (User, error) {
	sessionConfig := neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: databaseName,
	}
	session, err := driver.NewSession(sessionConfig)
	if err != nil {
		return User{}, err
	}
	defer session.Close()

	data := map[string]interface{}{"username": username}
	res, err := session.Run("MATCH (u:USER {username:$username}) RETURN u LIMIT 1", data)
	if err != nil {
		return User{}, err
	}

	for res.Next() {
		r := res.Record().GetByIndex(0).(neo4j.Node)
		user, err := ParseUser(r)
		if err != nil {
			continue
		}
		return user, nil
	}

	return User{}, fmt.Errorf("user not found")
}

func DeleteUser(username string) error {
	sessionConfig := neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: databaseName,
	}
	session, err := driver.NewSession(sessionConfig)
	if err != nil {
		return err
	}
	defer session.Close()

	data := map[string]interface{}{"username": username}
	_, err = session.Run("MATCH (user:USER {username:$username})-[r:MADE]->(url:URL) DETACH DELETE user DETACH DELETE url", data)
	return err
}

func VerifyUser(username, password string) bool {
	user, err := GetUser(username)
	if err != nil {
		return false
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

func AddUser(username, password string) error {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return err
	}

	sessionConfig := neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: databaseName,
	}
	session, err := driver.NewSession(sessionConfig)
	if err != nil {
		return err
	}
	defer session.Close()

	data := map[string]interface{}{"username": username, "password": string(hashedPass)}
	res, err := session.Run("CREATE (u:USER {username:$username, password:$password, created:datetime({ timezone: 'Europe/London' })})", data)
	if err != nil {
		return err
	}

	return res.Err()
}

func Link(username, shortened string) error {
	sessionConfig := neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: databaseName,
	}
	session, err := driver.NewSession(sessionConfig)
	if err != nil {
		return err
	}
	defer session.Close()

	data := map[string]interface{}{"username": username, "short": shortened}
	res, err := session.Run("MATCH (u:USER {username:$username}) MATCH (url:URL {short:$short}) CREATE (u)-[r:MADE]->(url)", data)
	if err != nil {
		return err
	}

	return res.Err()
}

func GetURLsOf(username string) ([]Record, error) {
	sessionConfig := neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: databaseName,
	}
	session, err := driver.NewSession(sessionConfig)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	data := map[string]interface{}{"username": username}
	res, err := session.Run("MATCH (u:URL)--(USER {username:$username}) RETURN u", data)
	if err != nil {
		return nil, err
	}

	var records []Record

	for res.Next() {
		r := res.Record().GetByIndex(0).(neo4j.Node)
		record, err := ParseRecord(r)
		if err != nil {
			continue
		}
		records = append(records, record)
	}

	return records, nil
}

func VerifyOwns(username, short string) bool {
	sessionConfig := neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: databaseName,
	}
	session, err := driver.NewSession(sessionConfig)
	if err != nil {
		return false
	}
	defer session.Close()

	data := map[string]interface{}{"username": username, "short": short}
	res, err := session.Run("MATCH (u:URL {short: $short})<-[r:MADE]-(USER {username:$username}) RETURN u", data)
	if err != nil {
		return false
	}

	return res.Next()
}

func DeleteURL(short string) error {
	sessionConfig := neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: databaseName,
	}
	session, err := driver.NewSession(sessionConfig)
	if err != nil {
		return err
	}
	defer session.Close()

	data := map[string]interface{}{"short": short}
	_, err = session.Run("MATCH (url:URL {short: $short}) DETACH DELETE url", data)
	return err
}

func ParseRecord(node neo4j.Node) (Record, error) {
	props := node.Props()

	short, ok := props["short"]
	if !ok {
		return Record{}, fmt.Errorf("short url not found")
	}
	long, ok := props["long"]
	if !ok {
		return Record{}, fmt.Errorf("long url not found")
	}
	created, ok := props["created"]
	if !ok {
		return Record{}, fmt.Errorf("created date not found")
	}

	return Record{
		Short:   short.(string),
		Long:    long.(string),
		Created: created.(time.Time),
	}, nil
}

func ParseUser(node neo4j.Node) (User, error) {
	props := node.Props()

	username, ok := props["username"]
	if !ok {
		return User{}, fmt.Errorf("username not found")
	}
	password, ok := props["password"]
	if !ok {
		return User{}, fmt.Errorf("password not found")
	}
	created, ok := props["created"]
	if !ok {
		return User{}, fmt.Errorf("created date not found")
	}

	return User{
		Username: username.(string),
		Password: password.(string),
		Created:  created.(time.Time),
	}, nil
}
