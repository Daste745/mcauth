package db

import (
	"database/sql"
	"fmt"
	"github.com/jinzhu/gorm"
	"log"
	"time"
)

// Config is a Postgres configuration.
type Config struct {
	Host               string        `yaml:"host"`
	Port               int           `yaml:"port"`
	User               string        `yaml:"username"`
	Password           string        `yaml:"password"`
	Database           string        `yaml:"database_name"`
	MaxConnections     int           `yaml:"max_connections"`
	MaxIdleConnections int           `yaml:"max_idle_connections"`
	ConnLifespan       time.Duration `yaml:"conn_lifespan"`
}

// Store is the database. For more information about each table
// visit their file. This is where they're all grouped together.
type Store struct {
	db    *sql.DB
	gDB   *gorm.DB
	Alts  AltsTable
	Auth  AuthTable
	Links LinksTable
}

const schema = "mcauth"

// GetStore returns the database and the structures that manage each table.
func GetStore(config Config) (c Store) {
	connConfig := fmt.Sprintf(
		"user=%s password=%s host=%s database=%s port=%d sslmode=disable",
		config.User, config.Password, config.Host, config.Database, config.Port,
	)
	gDB, err := gorm.Open("postgres", connConfig)

	if err != nil {
		log.Fatalln("Failed to connect to the postgres database\n", err.Error())
	}
	db := gDB.DB()

	gorm.DefaultTableNameHandler(gDB, schema+".")

	if err = db.Ping(); err != nil {
		log.Fatalln("Failed to ping the postgres database\n", err.Error())
	}

	if _, err := db.Exec("CREATE SCHEMA IF NOT EXISTS " + schema); err != nil {
		log.Fatalf("Failed to create schema \"%s\" because,\n%s", schema, err)
	}

	c = Store{
		db:  db,
		gDB: gDB,
	}

	c.db.SetMaxOpenConns(config.MaxConnections)
	c.db.SetMaxIdleConns(config.MaxIdleConnections)
	c.db.SetConnMaxLifetime(config.ConnLifespan)

	// Alt account management table
	c.Alts = GetAltsTable(gDB)
	// Authentication code table
	c.Auth = GetAuthTable(gDB)
	// Linked accounts table
	c.Links = GetLinksTable(gDB)

	return c
}
