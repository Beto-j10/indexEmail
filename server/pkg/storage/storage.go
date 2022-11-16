package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"server/config"
	def "server/pkg/definitions"
	"strings"
	"sync"
)

type Storage interface {
	Indexer()
	SearchMail(*def.Search) (*def.SearchResponse, error)
}

type storage struct {
	config *config.Config
}

func NewStorage(config *config.Config) Storage {
	return &storage{
		config: config,
	}
}

func (s *storage) SearchMail(search *def.Search) (*def.SearchResponse, error) {

	requestbody, err := json.Marshal(search)
	if err != nil {
		log.Printf("error marshalling search: %v", err)
		return nil, err
	}

	b := strings.Builder{}
	b.WriteString(s.config.Zinc.ZincHost)
	b.WriteString(s.config.Zinc.Target)
	b.WriteString(s.config.Zinc.Search)

	URL := b.String()

	client := &http.Client{}

	request, err := http.NewRequest("POST", URL, bytes.NewBuffer(requestbody))
	if err != nil {
		log.Printf("error creating request: %v", err)
		return nil, err
	}

	request.SetBasicAuth(s.config.Zinc.User, s.config.Zinc.Password)
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		log.Printf("error sending request: %v", err)
		return nil, err
	}

	defer response.Body.Close()

	resp := &def.SearchResponse{}

	err = json.NewDecoder(response.Body).Decode(resp)
	if err != nil {
		log.Printf("error decoding response: %v", err)
		return nil, err
	}

	fmt.Printf("Search results: %v \n", resp)

	return resp, nil

}

func (s *storage) Indexer() {
	files, err := os.ReadDir("./emails")
	if err != nil {
		log.Panicf("error reading directory: %+v", err)
	}

	c := make(chan int, 4)
	wg := sync.WaitGroup{}

	for _, file := range files {
		wg.Add(1)
		c <- 1
		go func(fileName string) {
			defer wg.Done()
			b := strings.Builder{}
			b.WriteString(s.config.Zinc.ZincHost)
			b.WriteString(s.config.Zinc.Target)
			b.WriteString(s.config.Zinc.DocCreate)

			URL := b.String()

			// requestBody, err := json.Marshal(emails)
			// if err != nil {
			// 	log.Panicf("error marshalling email: %v", err)
			// }

			file, err := os.Open("./emails/" + fileName)
			if err != nil {
				log.Panicf("error opening file: %v", err)
			}
			defer file.Close()

			client := &http.Client{}
			request, err := http.NewRequest("POST", URL, file)
			if err != nil {
				log.Panicf("error creating request: %v", err)
			}

			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			// request.Header.Set("Accept", "application/json")
			request.SetBasicAuth(s.config.Zinc.User, s.config.Zinc.Password)

			response, err := client.Do(request)
			if err != nil {
				log.Panicf("error sending request: %v", err)
			}
			defer response.Body.Close()

			fmt.Printf("Indexed %s with status %s \n", fileName, response.Status)
			<-c
		}(file.Name())

	}
	wg.Wait()

}
