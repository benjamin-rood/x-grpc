package main

import (
	"log"
	"os"
	"path/filepath"
)

// default use - just a glorified wrapper around a call to `os.OpenFile(...)`
type diskWriter struct {
	f            *os.File
	writeDirPath string
}

// Check interface conformity
var _ OpenWriteCloserLoader = &diskWriter{}

// uses the os package to open a file pointer so we can write bytes
// to a file on disk with the given filename
func (dw *diskWriter) Open(filename string) error {
	fp := dw.filePath(filename)
	log.Printf("opening file '%s'\n", fp)
	var err error
	dw.f, err = os.OpenFile(fp, os.O_RDWR|os.O_CREATE, 0644)
	return err
}

func (dw *diskWriter) Write(p []byte) (int, error) {
	return dw.f.Write(p)
}

func (dw *diskWriter) Close() error {
	log.Println("closing file")
	return dw.f.Close()
}

func (dw *diskWriter) Load(filename string) ([]byte, error) {
	return os.ReadFile(dw.filePath(filename))
}

func (dw *diskWriter) filePath(filename string) string {
	return filepath.Join(dw.writeDirPath, filename)
}
