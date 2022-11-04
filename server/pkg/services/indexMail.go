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
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
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

	// through all directories
	d, err := fs.ReadDir(fileSystem, root)
	if err != nil {
		log.Printf("Error reading root directory: %v", err)
		return err
	}

	var countFiles uint64
	c := make(chan int, 10)
	// var wg sync.WaitGroup
	eg := &errgroup.Group{}
	var countDir uint64
	countDir = uint64(len(d))

	for _, dir := range d {
		if dir.IsDir() {

			// through all subDirectories
			d, err := fs.ReadDir(fileSystem, root+"/"+dir.Name())
			if err != nil {
				log.Printf("Error reading directory: %v", err)
				return err
			}

			countDir += uint64(len(d))

			for _, subDir := range d {
				if subDir.IsDir() {
					fmt.Printf("\rProgress: %.1f%%", (float32(countFiles)/float32(totalFiles))*100)
					fileSystem := os.DirFS(dirPath + "/" + root + "/" + dir.Name() + "/" + subDir.Name())

					c <- 1
					// wg.Add(1)

					// Traverses a user's subfolders
					eg.Go(func() error {

						// defer wg.Done()
						emails := &email.EmailList{}

						err := fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
							if err != nil {
								return err
							}

							if d.IsDir() {
								if d.Name() != "." {
									// log.Printf("neted dir: %v route: %v", d.Name(), fileSystem)
									countDir++
								}

							} else {
								email, err := openFile(fileSystem, path, &countFiles)
								if err != nil {
									return err
								}
								emails.Emails = append(emails.Emails, *email)

							}

							return nil
						})
						if err != nil {
							log.Printf("Error walking directory: %v", err)
							<-c // set the semaphore free to send error. Otherwise the program will wait forever.
							return err
						}
						<-c
						// fmt.Printf("\n\n\nemails: %v", emails)
						err = i.storage.Indexer(emails)
						if err != nil {
							log.Printf("Error indexing emails: %v", err)
							return err
						}
						return nil
					})

				} /* else { //if no, there is file in user root
					fileSystem := os.DirFS(dirPath + "/" + root + "/" + dir.Name())
					email, err := openFile(fileSystem, subDir.Name(), &countFiles)
					if err != nil {
						return err
					}
					fmt.Printf("\rDoc: %v", email)
				} */
			}
		}
	}
	if err := eg.Wait(); err != nil {
		return err
	}

	// wg.Wait()

	log.Printf("\nTotal files: %v", countFiles)
	log.Printf("Total directories: %v", countDir)

	return nil
}

func openFile(fileSystem fs.FS, path string, countFiles *uint64) (*email.Email, error) {
	file, err := fileSystem.Open(path)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return nil, err
	}
	defer file.Close()

	isHeaderEmail := true
	isBodyEmail := false
	// headerType := ""
	fillEmail := &fillEmail{
		email: &email.Email{},
	}

	reader := bufio.NewReader(file)
	for {
		fillEmail.line, err = read(reader)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Error reading file: %v", err)
			return nil, err
		}
		if !isBodyEmail && len(fillEmail.line) == 0 {
			isBodyEmail = true
		}
		if isHeaderEmail {
			lineSize := len(fillEmail.line)
			var prefixLine string
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
			fillEmail.email.Body += string(fillEmail.line)
		}
	}

	atomic.AddUint64(countFiles, 1)
	return fillEmail.email, nil
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
	switch f.headerType {
	case "Date":
		f.email.Date += string(f.line[f.startByte:])
	case "From":
		f.email.From += string(f.line[f.startByte:])
	case "To":
		f.email.To += string(f.line[f.startByte:])
	case "Subject":
		f.email.Subject += string(f.line[f.startByte:])
	}
}

func countFiles(fileSystem fs.FS) (uint64, error) {
	start := time.Now()
	log.Printf("Counting files...")
	var countFiles uint64
	err := fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			if filepath.Ext(path) == "." {
				countFiles++
			}
		}

		return nil
	})
	if err != nil {
		log.Printf("Error walking directory: %v", err)

		return 0, err
	}
	elapsed := time.Since(start)
	log.Printf("Counted %v files in %v", countFiles, elapsed)
	return countFiles, nil
}
