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
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
)

type IndexService interface {
	IndexMail() error
}

type indexService struct {
	config config.Config
}

func NewIndexService(config *config.Config) IndexService {
	return &indexService{
		config: *config,
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

					// log.Printf("subDir: %v", subDir.Name())

					//TODO check performance
					// var fsysStr strings.Builder
					// fsysStr.WriteString(dirPath)
					// fsysStr.WriteString("/")
					// fsysStr.WriteString(root)
					// fsysStr.WriteString("/")
					// fsysStr.WriteString(dir.Name())
					// fsysStr.WriteString("/")
					// fsysStr.WriteString(subDir.Name())

					// fileSystem := os.DirFS(fsysStr.String())
					// progress in percent
					fmt.Printf("\rProgress: %.1f%%", (float32(countFiles)/float32(totalFiles))*100)
					fileSystem := os.DirFS(dirPath + "/" + root + "/" + dir.Name() + "/" + subDir.Name())

					c <- 1
					// wg.Add(1)

					// Traverses a user's subfolders
					eg.Go(func() error {

						// defer wg.Done()

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

								// check extension
								// if filepath.Ext(path) == "." {
								// 		// println(path)
								// 		// countFiles++
								// 	}
								err := openFile(fileSystem, path, &countFiles)
								if err != nil {
									return err
								}
							}

							return nil
						})
						if err != nil {
							log.Printf("Error walking directory: %v", err)
							<-c // set the semaphore free to send error. Otherwise the program will wait forever.
							return err
						}
						<-c
						return nil
					})

				} else { //if no, there is file in user root
					fileSystem := os.DirFS(dirPath + "/" + root + "/" + dir.Name())
					err := openFile(fileSystem, subDir.Name(), &countFiles)
					if err != nil {
						return err
					}
				}
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

func openFile(fileSystem fs.FS, path string, countFiles *uint64) error {
	file, err := fileSystem.Open(path)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		_, err := read(reader)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Error reading file: %v", err)
			return err
		}
		// log.Printf("Line: %v", string(line))
	}

	atomic.AddUint64(countFiles, 1)
	return nil
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
