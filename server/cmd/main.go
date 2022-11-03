package main

import (
	"fmt"
	"os"

	"server/config"
	"server/pkg/app"
	api "server/pkg/services"
	"server/pkg/storage"

	"github.com/go-chi/chi/v5"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Startup error: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	config, err := config.LoadConfig("./config.yml")
	if err != nil {
		return err
	}

	// create router
	router := chi.NewRouter()

	storage := storage.NewStorage()

	// create services
	indexService := api.NewIndexService(config, storage)

	// create server
	server := app.NewServer(router, indexService, config)

	// run server
	err = server.Run(config.Server.Port)
	if err != nil {
		return err
	}

	return nil

}
