package storage

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"server/config"
	"sync"
)

type Storage interface {
	Indexer()
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

// func (s *storage) Indexer(emails email.EmailList, c <-chan int) {
// 	URL := s.config.Zinc.ZincHost + s.config.Zinc.Target + s.config.Zinc.DocCreate
// 	requestBody, err := json.Marshal(emails)
// 	if err != nil {
// 		log.Panicf("error marshalling email: %v", err)
// 	}

// 	client := &http.Client{}
// 	request, err := http.NewRequest("POST", URL, bytes.NewBuffer(requestBody))
// 	if err != nil {
// 		log.Panicf("error creating request: %v", err)
// 	}

// 	request.Header.Set("Content-Type", "application/json")
// 	request.Header.Set("Accept", "application/json")
// 	request.SetBasicAuth(s.config.Zinc.User, s.config.Zinc.Password)

// 	_, err = client.Do(request)
// 	if err != nil {
// 		log.Panicf("error sending request: %v", err)
// 	}
// 	<-c
// }

// func (s *storage) Indexer() {
// 	files, err := os.ReadDir("./emails")
// 	if err != nil {
// 		log.Panicf("error reading directory: %v", err)
// 	}

// 	for _, file := range files {
// 		// if file.Name() == "index-final.json" {
// 		fmt.Println(file.Name())

// 		URL := s.config.Zinc.ZincHost + s.config.Zinc.DocCreate
// 		// requestBody, err := json.Marshal(emails)
// 		// if err != nil {
// 		// 	log.Panicf("error marshalling email: %v", err)
// 		// }

// 		file, err := os.Open("./emails/" + file.Name())
// 		if err != nil {
// 			log.Panicf("error opening file: %v", err)
// 		}

// 		client := &http.Client{}
// 		request, err := http.NewRequest("POST", URL, file)
// 		if err != nil {
// 			log.Panicf("error creating request: %v", err)
// 		}

// 		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
// 		// request.Header.Set("Accept", "application/json")
// 		request.SetBasicAuth(s.config.Zinc.User, s.config.Zinc.Password)

// 		response, err := client.Do(request)
// 		if err != nil {
// 			log.Panicf("error sending request: %v", err)
// 		}

// 		fmt.Println(response.Status)
// 		file.Close()
// 		response.Body.Close()

// 		// }
// 	}

// }

func (s *storage) SearchMail() error {
	return nil
}

func (s *storage) Indexer() {
	files, err := os.ReadDir("./emails")
	if err != nil {
		log.Panicf("error reading directory: %v", err)
	}

	c := make(chan int, 4)
	wg := sync.WaitGroup{}

	for _, file := range files {
		wg.Add(1)
		c <- 1
		go func(fileName string) {
			defer wg.Done()

			URL := s.config.Zinc.ZincHost + s.config.Zinc.DocCreate
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
