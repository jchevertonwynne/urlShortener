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
    Created time.Time
}

const (
    serverURL = "bolt://localhost:7687"
    databaseName = "neo4j"
    bcryptCost   = 10
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

func GetUrl(short string) (Record, error) {
    driver, err := neo4j.NewDriver(serverURL, neo4j.BasicAuth(dbUsername, dbPassword, ""), func(config *neo4j.Config) {
        config.Encrypted = false
    })
    if err != nil {
        return Record{}, err
    }
    defer driver.Close()

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
    res, err := session.Run("MATCH (u:URL {short:$short}) RETURN u.short AS short, u.long AS long, u.created AS created LIMIT 1", data)
    if err != nil {
        return Record{}, err
    }

    for res.Next() {
        long, ok := res.Record().Get("long")
        if !ok {
            continue
        }
        short, ok := res.Record().Get("short")
        if !ok {
            continue
        }
        created, ok := res.Record().Get("created")
        if !ok {
            continue
        }
        record := Record{
            Short:   short.(string),
            Long:    long.(string),
            Created: created.(neo4j.Date).Time(),
        }

        return record, nil
    }

    return Record{}, fmt.Errorf("url not found")
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
    res, err := session.Run("MATCH (u:USER {username:$username}) RETURN u.username AS username, u.password AS password, u.created AS created LIMIT 1", data)
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
            Created:  created.(neo4j.Date).Time(),
        }
        return user, nil
    }

    return User{}, fmt.Errorf("user not found")
}

func VerifyUser(username, password string) bool {
    user, err := GetUser(username)
    if err != nil {
        return false
    }
    err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
    if err != nil {
        return false
    }
    return true
}

func AddUser(username, password string) error {
    u, err := GetUser(username)
    if err == nil {
       fmt.Println(u)
       return fmt.Errorf("user %s already exists", username)
    }

    hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
    if err != nil {
        return err
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

    data := map[string]interface{}{"username": username, "password": string(hashedPass), "created": neo4j.DateOf(time.Now())}
    res, err := session.Run("CREATE (u:USER {username:$username, password:$password, created:$created})", data)
    if err != nil {
        return err
    }

    return res.Err()
}

func Link(username, shortened string) error {
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

    data := map[string]interface{}{"username": username, "short": shortened}
    res, err := session.Run("MATCH (u:USER {username:$username}) MATCH (url:URL {short:$short}) CREATE (u)-[r:MADE]->(url)", data)
    if err != nil {
        return err
    }

    return res.Err()
}

func GetURLsOf(username string) ([]Record, error) {
    driver, err := neo4j.NewDriver(serverURL, neo4j.BasicAuth(dbUsername, dbPassword, ""), func(config *neo4j.Config) {
        config.Encrypted = false
    })
    if err != nil {
        return nil, err
    }
    defer driver.Close()

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
    res, err := session.Run("MATCH (u:URL)--(user:USER {username:$username}) RETURN u.short AS short, u.long AS long, u.created AS created", data)
    if err != nil {
        return nil, err
    }

    var records []Record

    for res.Next() {
        r := res.Record()
        short, ok := r.Get("short")
        if !ok {
            continue
        }
        long, ok := r.Get("long")
        if !ok {
            continue
        }
        created, ok := r.Get("created")
        if !ok {
            continue
        }

        record := Record{
            Short:   short.(string),
            Long:    long.(string),
            Created: created.(neo4j.Date).Time(),
        }
        records = append(records, record)
    }

    return records, nil
}