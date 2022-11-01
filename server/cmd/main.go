package main

import (
	"fmt"
	"os"

	"server/config"
	"server/pkg/app"
	api "server/pkg/services"

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

	// create router
	router := chi.NewRouter()

	// create services
	indexService := api.NewIndexService(config)

	// create server
	server := app.NewServer(router, indexService, config)

	// run server
	err = server.Run(config.Server.Port)
	if err != nil {
		return err
	}

	return nil

}
