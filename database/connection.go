package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type connection struct {
	*sql.DB
}

type ConnectionParams struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

func GetDatabase(params *ConnectionParams) (*Database, error) {
	log.Debug("connecting to database")
	conURL := fmt.Sprintf("postgresql://%v:%v@%v:%v/%v?connect_timeout=1&sslmode=disable", params.User, params.Password, params.Host, params.Port, params.Name)
	con, err := sql.Open("postgres", conURL)
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %v", err)
	}

	if err := con.Ping(); err != nil {
		return nil, fmt.Errorf("could not connect to database: %v", err)
	}
	log.Debug("connected to database")

	d := &Database{
		Name:       params.Name,
		Schema:     "public",
		connection: &connection{con},
	}
	d.LoadState()
	return d, nil
}
