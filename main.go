package main

import (
	"log"
	"os"

	"duck/internal/cli"
)

func main() {
	app := cli.CreateApp()

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
