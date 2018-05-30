package main

import (
	"flag"
	"log"
	"sync"
)

func main() {
	var (
		parallel = flag.Int("parallel", 1, "parallel number of writing file")
		testDir  = flag.String("testDir", "testdata", "test data directory")
	)
	flag.Parse()
	log.Println("start checker")

	subDirs, err := getSubDirs(*testDir)
	if err != nil {
		log.Fatal(err)
	}

	pathCh := make(chan string)
	wg := sync.WaitGroup{}
	for i := 0; i < *parallel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				path, ok := <-pathCh
				if !ok {
					return
				}

				log.Printf("checking %s ...\n", path)
				c := NewChecker(path)
				if err := c.Verify(); err != nil {
					log.Println(err)
				}
			}
		}()
	}

	for _, path := range subDirs {
		pathCh <- path
	}
	close(pathCh)

	wg.Wait()

	log.Println("end checker")
}
