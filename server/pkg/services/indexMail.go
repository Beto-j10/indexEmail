package services

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"server/config"
	def "server/pkg/definitions"
	"server/pkg/storage"
	"strings"
	"sync"
	"time"
)

// IndexMail indexes all the mail in the maildir.
type IndexService interface {
	IndexMail() error
	Indexer() error
	SearchMail(query *def.Query) (*def.SearchResponse, error)
}

type indexService struct {
	config  config.Config
	storage storage.Storage
}

type fillEmail struct {
	email      *def.Email
	line       []byte
	headerType string
	startByte  int
}

func (f *fillEmail) reset() {
	f.email = &def.Email{}
	f.line = nil
	f.headerType = ""
	f.startByte = 0
}

func NewIndexService(config *config.Config, storage storage.Storage) IndexService {
	return &indexService{
		config:  *config,
		storage: storage,
	}
}

func (i *indexService) IndexMail() error {
	dirPath := i.config.Server.Dir
	root := "maildir"
	fileSystem := os.DirFS(dirPath)

	totalFiles, err := countFiles(fileSystem)
	if err != nil {
		return err
	}

	nWorkers := 4
	jobs := make(chan string, 50)
	results := make(chan *def.Email, 50)

	wg := sync.WaitGroup{}
	wg2 := sync.WaitGroup{}

	for w := 1; w <= nWorkers; w++ {
		wg.Add(1)
		go worker(jobs, results, &wg)
	}

	wg2.Add(1)
	go createFiles(results, &wg2)

	var fileCounter float64

	err = fs.WalkDir(fileSystem, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if filepath.Ext(path) == "." {
				fileCounter += 1
				jobs <- dirPath + path
			}
		}
		fmt.Printf("\rProgress: %.2f%% - files:%.0f / %.0f", ((fileCounter / totalFiles) * 100), fileCounter, totalFiles)
		return nil
	})
	if err != nil {
		log.Printf("Error walking directory: %v", err)
		return err
	}
	close(jobs)
	wg.Wait()
	close(results)
	wg2.Wait()

	// i.storage.Indexer()

	return nil
}

func worker(jobs <-chan string, result chan<- *def.Email, wg *sync.WaitGroup) {

	defer wg.Done()

	var (
		strLine       string
		isHeaderEmail bool = true
		isBodyEmail   bool = false

		isPrefix       bool = true
		errRead        error
		line, ln, body []byte

		b strings.Builder
	)

	fillEmail := &fillEmail{
		email: &def.Email{},
	}

	buff := make([]byte, 500*1024)

	for path := range jobs {
		file, err := os.Open(path)
		if err != nil {
			log.Panicf("Error opening file: %v", err)
		}

		isHeaderEmail = true
		isBodyEmail = false

		reader := bufio.NewReader(file)
		for {
			isPrefix = true
			line = nil
			ln = nil

			if isBodyEmail {
				n, err := reader.Read(buff)
				if n == 0 {
					if err != nil {
						if err == io.EOF {
							break
						}
						log.Panicf("Error reading file: %v", err)
					}
				}
				body = append(body, buff[:n]...)
				break

			} else { // isHeaderEmail

				for isPrefix && errRead == nil {
					line, isPrefix, errRead = reader.ReadLine()
					ln = append(ln, line...)
				}

				if errRead != nil {
					if errRead == io.EOF {
						break
					}
					log.Panicf("Error reading file: %v", errRead)
				}
			}

			if isHeaderEmail {

				fillEmail.line = ln
				strLine = string(ln)

				switch {
				case strings.HasPrefix(strLine, "Date:"):
					fillEmail.startByte = 5
					fillEmail.headerType = "Date"
					fillEmail.fillEmails(&b)
				case strings.HasPrefix(strLine, "From:"):
					fillEmail.startByte = 5
					fillEmail.headerType = "From"
					fillEmail.fillEmails(&b)
				case strings.HasPrefix(strLine, "To:"):
					fillEmail.startByte = 3
					fillEmail.headerType = "To"
					fillEmail.fillEmails(&b)
				case strings.HasPrefix(strLine, "Subject:"):
					fillEmail.startByte = 8
					fillEmail.headerType = "Subject"
					fillEmail.fillEmails(&b)
				case strings.HasPrefix(strLine, "Mime-Version:") || strings.HasPrefix(strLine, "Cc:"):
					fillEmail.startByte = 0
					fillEmail.headerType = ""
					isHeaderEmail = false
				default:
					fillEmail.startByte = 0
					fillEmail.fillEmails(&b)
				}
			} else if !isBodyEmail && len(ln) == 0 {
				isBodyEmail = true
			}
		}

		fillEmail.email.Body = string(body)

		result <- fillEmail.email

		// resets
		fillEmail.reset()
		buff = make([]byte, 500*1024)
		body = nil

		file.Close()

	}

}

func (f *fillEmail) fillEmails(b *strings.Builder) {

	switch f.headerType {
	case "Date":
		b.WriteString(f.email.Date)
		b.Write(f.line[f.startByte:])
		f.email.Date = b.String()
		b.Reset()
	case "From":
		b.WriteString(f.email.From)
		b.Write(f.line[f.startByte:])
		f.email.From = b.String()
		b.Reset()
	case "To":
		b.WriteString(f.email.To)
		b.Write(f.line[f.startByte:])
		f.email.To = b.String()
		b.Reset()
	case "Subject":
		b.WriteString(f.email.Subject)
		b.Write(f.line[f.startByte:])
		f.email.Subject = b.String()
		b.Reset()
	}
}

func createFiles(result <-chan *def.Email, wg *sync.WaitGroup) {

	defer wg.Done()

	emailList := &def.EmailList{}
	counter := 0
	batch := 0
	c := make(chan int, 5)
	for emailResult := range result {
		emailList.Emails = append(emailList.Emails, *emailResult)
		counter++
		if counter == 5000 {

			batch++
			name := fmt.Sprintf("./emails/index-%d.json", batch)
			// save to db
			c <- 1
			go func(emailList def.EmailList) {

				b, err := json.Marshal(emailList)
				if err != nil {
					log.Panicf("Error marshalling to json: %v", err)
				}

				os.WriteFile(name, b, 0644)

				<-c
			}(*emailList)
			emailList.Emails = nil

			counter = 0
		}
	}

	b, err := json.Marshal(emailList)
	if err != nil {
		log.Panicf("Error marshalling to json: %v", err)
	}
	os.WriteFile("./emails/index-final.json", b, 0644)

}

func countFiles(fileSystem fs.FS) (float64, error) {
	start := time.Now()
	log.Printf("Counting files...")
	var fileCounter float64
	err := fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			if filepath.Ext(path) == "." {
				fileCounter++
			}
		}

		return nil
	})
	if err != nil {
		log.Printf("Error walking directory: %v", err)

		return 0, err
	}
	elapsed := time.Since(start)
	log.Printf("Counted %0.f files in %v", fileCounter, elapsed)
	return fileCounter, nil
}
