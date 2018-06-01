package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const (
	KiB = 1024
	MiB = 1024 * KiB

	MinFileSize = 4 * KiB
	MaxFileSize = 5 * MiB

	NumberOfFilesInDir  = 1000
	SizeOfSyncFileRange = 1 * MiB
)

func main() {
	var (
		parallel       = flag.Int("parallel", 1, "parallel number of writing file")
		testDir        = flag.String("testDir", "testdata", "test data directory")
		numFiles       = flag.Int("numFiles", NumberOfFilesInDir, "number of file in a directory")
		syncFileRange  = flag.Int("syncFileRange", SizeOfSyncFileRange, "size of sync_file_range")
		syncClose      = flag.Bool("syncClose", false, "call sync_file_range before close")
		disableFadvice = flag.Bool("disableFadvice", false, "do not call fadvice after sync_file_range")
	)
	flag.Parse()

	log.Printf("size of syncFileRange: %d", *syncFileRange)

	wg := sync.WaitGroup{}
	for i := 0; i < *parallel; i++ {
		wg.Add(1)

		dir := fmt.Sprintf("%s/sub%04d", *testDir, i)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatal(err)
		}

		go func(dir string) {
			defer wg.Done()

			log.Printf("start writing files into %s ...\n", dir)
			for i, size := range genRandomSizes(*numFiles, MinFileSize, MaxFileSize) {
				path := filepath.Join(dir, fmt.Sprintf("%03d.data", i))
				if err := writeFile(path, size, *syncFileRange, *syncClose, *disableFadvice); err != nil {
					fmt.Println(err) // ignore error
				}
			}
			log.Printf("end writing files into %s\n", dir)
		}(dir)
	}

	wg.Wait()
}
