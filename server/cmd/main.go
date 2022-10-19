package main

import (
	"flag"
	"fmt"
	"os"

	"server/config"
	"server/pkg/api"
	"server/pkg/app"

	"github.com/go-chi/chi/v5"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Startup error: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	config, err := config.LoadConfig("../config.yml")
	if err != nil {
		return err
	}

	serverPort := flag.String("port", config.Server.Port, "server port")
	flag.Parse()

	// create router
	router := chi.NewRouter()

	// create services
	indexService := api.NewIndexService()

	// create server
	server := app.NewServer(router, indexService)

	// run server
	err = server.Run(*serverPort)
	if err != nil {
		return err
	}

	return nil

}
