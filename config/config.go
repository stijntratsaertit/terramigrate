package config

import (
	"os"
	"stijntratsaertit/terramigrate/database"
	"strconv"

	log "github.com/sirupsen/logrus"
)

var (
	databaseHost     = os.Getenv("DATABASE_HOST")
	databasePort     = os.Getenv("DATABASE_PORT")
	databaseUser     = os.Getenv("DATABASE_USER")
	databasePassword = os.Getenv("DATABASE_PASSWORD")
	databaseName     = os.Getenv("DATABASE_NAME")
)

type Config struct {
	databaseHost     string
	databasePort     int
	databaseUser     string
	databasePassword string
	databaseName     string
}

func Get() *Config {
	port := 5432
	if databasePort != "" {
		var err error
		port, err = strconv.Atoi(databasePort)
		if err != nil {
			log.Errorf("could not parse database port: %v", err)
		}
	}

	return &Config{
		databaseHost:     databaseHost,
		databasePort:     port,
		databaseUser:     databaseUser,
		databasePassword: databasePassword,
		databaseName:     databaseName,
	}
}

func (c *Config) DatabaseConnectionParams() *database.ConnectionParams {
	return &database.ConnectionParams{
		Host:     c.databaseHost,
		Port:     c.databasePort,
		User:     c.databaseUser,
		Password: c.databasePassword,
		Name:     c.databaseName,
	}
}
