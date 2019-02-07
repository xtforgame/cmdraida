package tests

import (
	"os"
)

type FileWriter struct {
	file *os.File
}

func (fw *FileWriter) Write(p []byte) (int, error) {
	i, e := fw.file.Write(p)
	if e == nil && i > 0 {
		fw.file.Sync()
	}
	return i, e
}

func (fw *FileWriter) Close() error {
	if fw.file != nil {
		return fw.file.Close()
	}
	return nil
}
