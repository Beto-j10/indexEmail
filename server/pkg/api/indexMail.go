package api

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type IndexService interface {
	IndexMail() error
}

type indexService struct {
}

func NewIndexService() IndexService {
	return &indexService{}
}

func (s *indexService) IndexMail() error {
	root := "<path>"
	fileSystem := os.DirFS(root)

	countFiles := 0
	i := 1
	dirName := ""

	err := fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if i < 20 {
			if d.IsDir() {
				log.Printf("Dir: %v", path)
				dirName = d.Name()
			} else {

				if filepath.Ext(path) == "." {
					// 		// println(path)
					// 		// countFiles++
					// 	}

					// file to json
					file, err := fileSystem.Open(path)
					if err != nil {
						log.Printf("Error opening file: %v", err)
						return err
					}
					defer file.Close()

					scanner := bufio.NewScanner(file)
					for scanner.Scan() {
						line := scanner.Text()
						if strings.HasPrefix(line, "To:") {
							fmt.Println(line)
						}
					}

					if err := scanner.Err(); err != nil {
						log.Printf("Error scanning file: %v", err)
						return err
					}

					log.Printf("File: %v dirName:%v", path, dirName)
					countFiles++
				}
			}
			i++
		}

		// if !d.IsDir() {
		// 	if filepath.Ext(path) == "." {
		// 		// println(path)
		// 		// countFiles++
		// 	}
		// }
		return nil

	})
	if err != nil {
		log.Printf("Error walking directory: %v", err)
		return err
	}
	println(countFiles)

	return nil
}
