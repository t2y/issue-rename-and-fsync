package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/xattr"
)

func getChecksum(path string) ([]byte, error) {
	vv, err := xattr.Get(path, "user.md5")
	if err != nil {
		if e, ok := err.(*xattr.Error); ok {
			return nil, e.Err
		}
		return nil, err
	}

	checksum := make([]byte, hex.DecodedLen(len(vv)))
	n, err := hex.Decode(checksum, vv)
	if err != nil {
		return nil, fmt.Errorf("checksum is invalid encoding: %s", path)
	}
	if n != md5.Size {
		return nil, fmt.Errorf("checksum is invalid length: %s", path)
	}
	return checksum, nil
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
