package services

import (
	"bufio"
	"io"
	"io/fs"
	"log"
	"os"
	"server/config"
	"sync/atomic"

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
	// through all directories
	d, err := fs.ReadDir(fileSystem, root)
	if err != nil {
		log.Printf("Error reading root directory: %v", err)
		// if(strings.HasSuffix(err.Error(), "no such file or directory")) {
		// 	log.Printf("Error reading directory: %v", err)
		// }
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
			// log.Printf("Dir: %v", dir.Name())
			// through all directories
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
					fileSystem := os.DirFS(dirPath + "/" + root + "/" + dir.Name() + "/" + subDir.Name())

					c <- 1
					// wg.Add(1)
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
								// log.Printf("File: %v", path)ss

								// check extension
								// if filepath.Ext(path) == "." {
								// 		// println(path)
								// 		// countFiles++
								// 	}

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

								// log.Printf("File: %v dirName:%v", path, w.dirName)
								// w.countFiles++
								atomic.AddUint64(&countFiles, 1)

								// }
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

				}
			}
		}
	}
	if err := eg.Wait(); err != nil {
		return err
	}

	// wg.Wait()

	log.Printf("Total files: %v", countFiles)
	log.Printf("Total directories: %v", countDir)

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
