package main

import (
	"fmt"
	"os"

	"server/config"
	"server/pkg/services"
	"server/pkg/storage"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error processing files: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	config, err := config.LoadConfig("./config.yml")
	if err != nil {
		return err
	}

	// create storage
	storage := storage.NewStorage(config)

	// create services
	indexService := services.NewIndexService(config, storage)
	err = indexService.IndexMail()
	if err != nil {
		return err
	}

	return nil

}
