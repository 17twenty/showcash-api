package main

import (
	"flag"
	"log"

	"github.com/17twenty/showcash-api"
)

func main() {
	log.Println("Starting Showcash API")

	flag.Parse()
	log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)

	c := showcash.New()
	c.Start()
}
