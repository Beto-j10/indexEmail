package services

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"server/config"
	"server/pkg/email"
	"server/pkg/storage"
	"strings"
	"time"
)

// IndexMail indexes all the mail in the maildir.
type IndexService interface {
	IndexMail() error
}

type indexService struct {
	config  config.Config
	storage storage.Storage
}

type fillEmail struct {
	email      *email.Email
	line       []byte
	headerType string
	startByte  int
}

func (f *fillEmail) reset() {
	f.email = &email.Email{}
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

	nWorkers := 10
	jobs := make(chan string, 50)
	results := make(chan *email.Email, 50)
	exit := make(chan bool)

	for w := 1; w <= nWorkers; w++ {
		go worker(jobs, results)
	}

	go indexer(results, exit, i)

	var fileCounter uint64
	// emails := &email.EmailList{}
	err = fs.WalkDir(fileSystem, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fmt.Printf("\rProgress: %.1f%%", (float32(fileCounter)/float32(totalFiles))*100)
		if !d.IsDir() {
			if filepath.Ext(path) == "." {
				fileCounter++
				jobs <- dirPath + path
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("Error walking directory: %v", err)
		return err
	}
	close(jobs)
	exit <- true
	return nil
}

func worker(jobs <-chan string, result chan<- *email.Email) {

	var (
		b             strings.Builder
		lineSize      int
		prefixLine    string
		isHeaderEmail bool       = true
		isBodyEmail   bool       = false
		fillEmail     *fillEmail = &fillEmail{
			email: &email.Email{},
		}
	)

	for path := range jobs {
		file, err := os.Open(path)
		if err != nil {
			log.Panicf("Error opening file: %v", err)
		}

		isHeaderEmail = true
		isBodyEmail = false
		// headerType := ""

		reader := bufio.NewReader(file)
		for {
			fillEmail.line, err = read(reader)
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Panicf("Error reading file: %v", err)
			}
			if !isBodyEmail && len(fillEmail.line) == 0 {
				isBodyEmail = true
			}
			if isHeaderEmail {
				lineSize = len(fillEmail.line)
				if lineSize >= 13 {
					prefixLine = string(fillEmail.line[:13])
				} else {
					prefixLine = string(fillEmail.line[:lineSize])
				}
				switch {
				case strings.HasPrefix(prefixLine, "Date:"):
					fillEmail.startByte = 5
					fillEmail.headerType = "Date"
					fillEmail.fillEmails()
				case strings.HasPrefix(prefixLine, "From:"):
					fillEmail.startByte = 5
					fillEmail.headerType = "From"
					fillEmail.fillEmails()
				case strings.HasPrefix(prefixLine, "To:"):
					fillEmail.startByte = 3
					fillEmail.headerType = "To"
					fillEmail.fillEmails()
				case strings.HasPrefix(prefixLine, "Subject:"):
					fillEmail.startByte = 8
					fillEmail.headerType = "Subject"
					fillEmail.fillEmails()
				case strings.HasPrefix(prefixLine, "Mime-Version:") || strings.HasPrefix(prefixLine, "Cc:"):
					fillEmail.startByte = 0
					fillEmail.headerType = ""
					isHeaderEmail = false
				default:
					fillEmail.startByte = 0
					fillEmail.fillEmails()
				}
			}
			if isBodyEmail {
				b.WriteString(fillEmail.email.Body)
				b.WriteString(string(fillEmail.line))
				fillEmail.email.Body = b.String()
				b.Reset()
				// fillEmail.email.Body += string(fillEmail.line)
			}
		}

		// atomic.AddUint64(fileCounter, 1)
		result <- fillEmail.email
		fillEmail.reset()
		file.Close()

	}
}

func read(r *bufio.Reader) ([]byte, error) {
	var (
		isPrefix bool  = true
		err      error = nil
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return ln, err
}

func (f *fillEmail) fillEmails() {
	// var b strings.Builder

	switch f.headerType {
	case "Date":
		// b.WriteString(f.email.Date)
		// b.WriteString(string(f.line[f.startByte:]))
		// f.email.Date = b.String()
		f.email.Date += string(f.line[f.startByte:])
	case "From":
		// b.WriteString(f.email.From)
		// b.WriteString(string(f.line[f.startByte:]))
		// f.email.From = b.String()
		f.email.From += string(f.line[f.startByte:])
	case "To":
		// b.WriteString(f.email.To)
		// b.WriteString(string(f.line[f.startByte:]))
		// f.email.To = b.String()
		f.email.To += string(f.line[f.startByte:])
	case "Subject":
		// b.WriteString(f.email.Subject)
		// b.WriteString(string(f.line[f.startByte:]))
		// f.email.Subject = b.String()
		f.email.Subject += string(f.line[f.startByte:])
	}
}

func indexer(result <-chan *email.Email, exit <-chan bool, i *indexService) {
	emailList := &email.EmailList{}
	counter := 0
	batch := 0
	for email := range result {
		emailList.Emails = append(emailList.Emails, *email)
		counter++
		if counter == 5000 {
			batch++
			// save to db
			i.storage.Indexer(emailList)
			fmt.Printf("\rSaving to db %d\n", batch)
			emailList.Emails = nil
			counter = 0
		}
	}
	i.storage.Indexer(emailList)
	fmt.Printf("\nSaving to db  out #######%d\n", len(emailList.Emails))
	<-exit
}

func countFiles(fileSystem fs.FS) (uint64, error) {
	start := time.Now()
	log.Printf("Counting files...")
	var fileCounter uint64
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
	log.Printf("Counted %v files in %v", fileCounter, elapsed)
	return fileCounter, nil
}
