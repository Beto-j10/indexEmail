package storage

import (
	"server/pkg/email"
)

type Storage interface {
	Indexer(*email.Email) error
	SearchMail() error
}

type storage struct {
}

func NewStorage() Storage {
	return &storage{}
}

func (s *storage) Indexer(email *email.Email) error {
	return nil
}

func (s *storage) SearchMail() error {
	return nil
}
