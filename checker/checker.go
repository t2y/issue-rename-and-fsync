package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
)

type Checker struct {
	path string
}

func NewChecker(path string) *Checker {
	return &Checker{
		path: path,
	}
}

func (c *Checker) Verify() error {
	paths, err := walkDir(c.path)
	if err != nil {
		return err
	}

	for _, path := range paths {
		storedChecksum, err := getChecksum(path)
		if err != nil {
			fmt.Println("cannot get checksum from xattr")
			fmt.Printf("  - path: %s\n", path)
			continue
		}

		f, err := os.Open(path)
		if err != nil {
			log.Println(err)
			continue
		}

		h := md5.New()
		if _, err := io.Copy(h, f); err != nil {
			log.Println(err)
			continue
		}

		if calc := h.Sum(nil); !bytes.Equal(calc, storedChecksum) {
			fmt.Println("checksum is mismatched")
			fmt.Printf("  - path: %s\n", path)
			fmt.Printf("  - user.md5:   %x\n", storedChecksum)
			fmt.Printf("  - calculated: %x\n", calc)
			if info, err := f.Stat(); err == nil {
				if storedSize, err := getFileSize(path); err == nil {
					fmt.Printf("  - user.size  : %d\n", storedSize)
				}
				fmt.Printf("  - actual size: %d\n", info.Size())
				fmt.Printf("  - mtime: %v\n", info.ModTime())
			}
			fmt.Println()
		}
	}
	return nil
}
