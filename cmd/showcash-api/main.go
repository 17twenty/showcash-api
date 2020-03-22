package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/17twenty/showcash-api"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq" // this is for the DAO
)

func main() {
	log.Println("Starting Showcash API")

	log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)
	isStrict12FA := flag.Bool("strict", false, "Calls os.Exit(1) if required ENV vars is not set")
	flag.Parse()
	config := loadConfig(isStrict12FA)

	// Get the execution directory of the binary
	// Exexute migrations
	{
		ex, err := os.Executable()
		if err != nil {
			log.Fatalln("Can't find binary!", err)
		}
		exPath := filepath.Dir(ex)
		migrationDir := fmt.Sprintf("file:///%s/migrations", exPath)
		log.Println("Prepping migrations from:", migrationDir)

		// Setup the database
		m, err := migrate.New(
			migrationDir,
			fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
				config.Database.User,
				config.Database.Password,
				config.Database.Host,
				config.Database.Port,
				config.Database.Name,
			),
		)
		if err != nil {
			log.Fatalln("Couldn't create database connection", err)
		}
		if err := m.Up(); !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalln("Couldn't Up()", err)
		} else {
			log.Println("Migration not required. All good.")
		}
	}

	dao, err := showcash.NewDAO(
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Name,
	)

	if err != nil || !dao.IsConnected() {
		log.Fatalln("Couldn't open database -", err)
	}

	c := showcash.New(
		dao,
		config.UseS3,
	)
	c.Start()
}
