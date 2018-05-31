package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pkg/xattr"
)

func getChecksum(path string) ([]byte, error) {
	value, err := xattr.Get(path, "user.md5")
	if err != nil {
		if e, ok := err.(*xattr.Error); ok {
			return nil, e.Err
		}
		return nil, err
	}

	checksum := make([]byte, hex.DecodedLen(len(value)))
	n, err := hex.Decode(checksum, value)
	if err != nil {
		return nil, fmt.Errorf("checksum is invalid encoding: %s", path)
	}
	if n != md5.Size {
		return nil, fmt.Errorf("checksum is invalid length: %s", path)
	}
	return checksum, nil
}

func getFileSize(path string) (int, error) {
	size, err := xattr.Get(path, "user.size")
	if err != nil {
		if e, ok := err.(*xattr.Error); ok {
			return 0, e.Err
		}
		return 0, err
	}
	return strconv.Atoi(string(size))
}

func getSubDirs(baseDir string) ([]string, error) {
	subDirs := make([]string, 0)
	err := filepath.Walk(
		baseDir,
		func(path string, f os.FileInfo, err error) error {
			if f.IsDir() && path != baseDir {
				subDirs = append(subDirs, path)
			}
			return nil
		},
	)
	return subDirs, err
}

func walkDir(baseDir string) ([]string, error) {
	paths := make([]string, 0)
	err := filepath.Walk(
		baseDir,
		func(path string, f os.FileInfo, err error) error {
			if os.IsNotExist(err) {
				return nil
			}

			if err != nil {
				return err
			}

			if f.IsDir() {
				return nil
			}

			paths = append(paths, path)
			return nil
		},
	)
	return paths, err
}
