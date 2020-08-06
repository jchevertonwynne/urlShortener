package database

import (
    "fmt"
    "github.com/neo4j/neo4j-go-driver/neo4j"
    "time"
)

type Record struct {
    Shortened string
    Full      string
    Created   time.Time
}

type User struct {
    Username string
    Password string
    Created time.Time
}

const (
    serverURL = "bolt://localhost:7687"
    databaseName = "neo4j"
)

var (
    dbUsername string
    dbPassword string
)

func Init(username, password string) {
    dbUsername = username
    dbPassword = password
}

func AddURL(long, short string) error {
    driver, err := neo4j.NewDriver(serverURL, neo4j.BasicAuth(dbUsername, dbPassword, ""), func(config *neo4j.Config) {
        config.Encrypted = false
    })
    if err != nil {
        return err
    }
    defer driver.Close()

    sessionConfig := neo4j.SessionConfig{
        AccessMode:   neo4j.AccessModeWrite,
        DatabaseName: databaseName,
    }
    session, err := driver.NewSession(sessionConfig)
    if err != nil {
        return err
    }
    defer session.Close()

    data := map[string]interface{}{"long": long, "short": short, "created": neo4j.DateOf(time.Now())}
    res, err := session.Run("CREATE (u:URL {long:$long, short:$short, created:$created})", data)
    if err != nil {
        return err
    }

    return res.Err()
}

func GetUrl(short string) (string, error) {
    driver, err := neo4j.NewDriver(serverURL, neo4j.BasicAuth(dbUsername, dbPassword, ""), func(config *neo4j.Config) {
        config.Encrypted = false
    })
    if err != nil {
        return "", err
    }
    defer driver.Close()

    sessionConfig := neo4j.SessionConfig{
        AccessMode:   neo4j.AccessModeRead,
        DatabaseName: databaseName,
    }
    session, err := driver.NewSession(sessionConfig)
    if err != nil {
        return "", err
    }
    defer session.Close()

    data := map[string]interface{}{"short": short}
    res, err := session.Run("MATCH (u:URL {short:$short}) RETURN u.short as short, u.long as long LIMIT 1", data)
    if err != nil {
        return "", err
    }

    for res.Next() {
        got, ok := res.Record().Get("long")
        if !ok {
            continue
        }

        return got.(string), nil
    }

    return "", fmt.Errorf("url not found")
}

func GetUser(username string) (User, error) {
    driver, err := neo4j.NewDriver(serverURL, neo4j.BasicAuth(dbUsername, dbPassword, ""), func(config *neo4j.Config) {
        config.Encrypted = false
    })
    if err != nil {
        return User{}, err
    }
    defer driver.Close()

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
    res, err := session.Run("MATCH (u:USER {username:$username}) RETURN u.username as username, u.password as password, u.created as created LIMIT 1", data)
    if err != nil {
        return User{}, err
    }

    for res.Next() {
        r := res.Record()
        username, ok := r.Get("username")
        if !ok {
            continue
        }
        password, ok := r.Get("password")
        if !ok {
            continue
        }
        created, ok := r.Get("created")
        if !ok {
            continue
        }

        user := User{
            Username: username.(string),
            Password: password.(string),
            Created: created.(neo4j.Date).Time(),
        }
        return user, nil
    }

    return User{}, fmt.Errorf("user not found")
}

func AddUser(username, password string) error {
    u, err := GetUser(username)
    if err == nil {
       fmt.Println(u)
       return fmt.Errorf("user %s already exists", username)
    }

    driver, err := neo4j.NewDriver(serverURL, neo4j.BasicAuth(dbUsername, dbPassword, ""), func(config *neo4j.Config) {
        config.Encrypted = false
    })
    if err != nil {
        return err
    }
    defer driver.Close()

    sessionConfig := neo4j.SessionConfig{
        AccessMode:   neo4j.AccessModeWrite,
        DatabaseName: databaseName,
    }
    session, err := driver.NewSession(sessionConfig)
    if err != nil {
        return err
    }
    defer session.Close()

    data := map[string]interface{}{"username": username, "password": password, "created": neo4j.DateOf(time.Now())}
    res, err := session.Run("CREATE (u:USER {username:$username, password:$password, created:$created})", data)
    if err != nil {
        return err
    }

    return res.Err()
}