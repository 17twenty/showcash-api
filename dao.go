package showcash

import (
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
)

type DAO struct {
	db *sqlx.DB
}

// NewDAO is a postgres backed version of the DAO
func NewDAO(username, password string, host string, port int, database string) (*DAO, error) {
	uri := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		username,
		password,
		host,
		port,
		database,
	)
	d := &DAO{}

	var err error
	d.db, err = sqlx.Open("postgres", uri)
	if err != nil {
		log.Println("Couldn't Open() DB with:",
			username,
			"******",
			host,
			port,
			database,
			err)
		return nil, errBadDAO
	}

	if !d.IsConnected() {
		log.Println("Couldn't Ping() DB with:",
			username,
			"******",
			host,
			port,
			database,
			err)
		return nil, errBadDAO
	}
	// Stops us having to add the db tags
	d.db.Mapper = reflectx.NewMapperFunc("json", strings.ToLower)
	return d, nil
}

// IsConnected is a healthcheck for the DAO
// In memory versions would be static but DB backed would ping()
func (d *DAO) IsConnected() bool {
	if d == nil {
		return false
	}
	if err := d.db.Ping(); err != nil {
		log.Println("Error from database ", err)
		return false
	}
	return true
}
