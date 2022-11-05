package storage

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"server/config"
	"server/pkg/email"
)

type Storage interface {
	Indexer(*email.EmailList)
	SearchMail() error
}

type storage struct {
	config *config.Config
}

func NewStorage(config *config.Config) Storage {
	return &storage{
		config: config,
	}
}

func (s *storage) Indexer(emails *email.EmailList) {
	URL := s.config.Zinc.ZincHost + s.config.Zinc.Target + s.config.Zinc.DocCreate
	requestBody, err := json.Marshal(emails)
	if err != nil {
		log.Panicf("error marshalling email: %v", err)
	}

	client := &http.Client{}
	request, err := http.NewRequest("POST", URL, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Panicf("error creating request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.SetBasicAuth(s.config.Zinc.User, s.config.Zinc.Password)

	_, err = client.Do(request)
	if err != nil {
		log.Panicf("error sending request: %v", err)
	}
}

func (s *storage) SearchMail() error {
	return nil
}
