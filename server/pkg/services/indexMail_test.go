package services

import (
	"log"
	"os"
	"server/config"
	def "server/pkg/definitions"
	"testing"

	"github.com/joho/godotenv"
)

type mockStorage struct{}

func (m *mockStorage) SearchMail(*def.Search) (*def.SearchResponse, error) {
	return nil, nil
}

func (m *mockStorage) Indexer() {}

func TestIndexEmail(t *testing.T) {
	err := godotenv.Load("../../test.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	tempDir := t.TempDir()

	dirBatch := tempDir + os.Getenv("TEST_NAME_BATCH")
	dirFinalBatch := tempDir + os.Getenv("TEST_NAME_FINAL_BATCH")

	mockDB := &mockStorage{}

	config := &config.Config{
		Server: config.ServerConfig{
			Dir:           os.Getenv("TEST_MAILDIR"),
			DirRootBatch:  tempDir,
			DirBatch:      dirBatch,
			DirFinalBatch: dirFinalBatch,
		},
	}
	indexService := NewIndexService(config, mockDB)
	err = indexService.IndexMail()
	if err != nil {
		t.Errorf("error creating files, %v", err)
	}

}
