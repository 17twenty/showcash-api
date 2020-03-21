package main

import (
	"flag"
	"log"

	"github.com/17twenty/showcash-api"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	log.Println("Starting Showcash API")

	log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)
	isStrict12FA := flag.Bool("strict", false, "Calls os.Exit(1) if required ENV vars is not set")
	flag.Parse()
	config := loadConfig(isStrict12FA)

	// Setup the database
	m, err := migrate.New(
		"file://db/migrations",
		"postgres://postgres:postgres@localhost:5432/example?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil {
		log.Fatal(err)
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
