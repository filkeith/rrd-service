package main

import (
	"log"

	"aerospike.com/rrd/internal"
)

func main() {
	app, err := internal.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	if err = app.Start(); err != nil {
		log.Fatal(err)
	}
}
