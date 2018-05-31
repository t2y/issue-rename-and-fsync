package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pkg/xattr"
)

const ()

type Writer struct {
	path          string
	tmp           *os.File
	h             hash.Hash
	checksum      []byte
	size          int
	syncFileRange int64
	syncOffset    int64
	syncLen       int64
}

func NewWriter(path string, size int, syncFileRange int) (w *Writer, err error) {
	tmp, err := ioutil.TempFile(filepath.Dir(path), "tmp-"+filepath.Base(path)+"-")
	if err != nil {
		return nil, errors.New("create temp file")
	}

	w = &Writer{
		path:          path,
		tmp:           tmp,
		h:             md5.New(),
		size:          size,
		syncFileRange: int64(syncFileRange),
	}
	return
}

func (w *Writer) Write(p []byte) (int, error) {
	n, err := w.tmp.Write(p)
	w.h.Write(p[:n])
	if err != nil {
		return n, err
	}
	w.syncLen += int64(n)

	if w.syncFileRange <= 0 {
		return n, nil
	}

	if w.syncLen >= w.syncFileRange {
		if err := w.sync(); err != nil {
			return n, err
		}
	}
	return n, nil
}

func (w *Writer) sync() error {
	if err := syncFileRange(int(w.tmp.Fd()), w.syncOffset, w.syncLen, SYNC_FILE_RANGE_WRITE); err != nil {
		return errors.New("sync_file_range")
	}

	if err := fadvise(int(w.tmp.Fd()), w.syncOffset, w.syncLen, FADV_DONTNEED); err != nil {
		return errors.New("posix_fadvise")
	}

	w.syncOffset, w.syncLen = w.syncOffset+w.syncLen, 0
	return nil
}

func (w *Writer) Rollback() error {
	w.tmp.Close()
	return os.Remove(w.tmp.Name())
}

func (w *Writer) Commit() error {
	w.checksum = w.h.Sum(nil)
	if err := xattr.Set(w.tmp.Name(), "user.md5", []byte(hex.EncodeToString(w.checksum))); err != nil {
		return errors.New("set user.md xattr")
	}
	if err := xattr.Set(w.tmp.Name(), "user.size", []byte(strconv.Itoa(w.size))); err != nil {
		return errors.New("set user.size xattr")
	}

	if err := w.tmp.Close(); err != nil {
		return errors.New("close temp file")
	}
	if err := os.Rename(w.tmp.Name(), w.path); err != nil {
		return errors.New("rename temp file")
	}
	return nil
}

func writeFile(path string, size, syncFileRange int) error {
	w, err := NewWriter(path, size, syncFileRange)
	if err != nil {
		return err
	}

	if _, err := io.CopyN(w, genData(size), int64(size)); err != nil {
		w.Rollback()
		return err
	}

	if err := w.Commit(); err != nil {
		w.Rollback()
		return err
	}

	return nil
}
