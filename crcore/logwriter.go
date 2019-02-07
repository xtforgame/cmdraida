package crcore

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
)

type LogWriter struct {
	_type  string
	owner  *Reporter
	file   *os.File
	buffer bytes.Buffer
	writer io.Writer
}

func NewLogWriter(_type string, owner *Reporter) *LogWriter {
	return &LogWriter{
		_type:  _type,
		owner:  owner,
		buffer: bytes.Buffer{},
	}
}

func GetLogPath(path, _type string) string {
	return filepath.Join(path, _type+".log")
}

func (lw *LogWriter) Buffer() bytes.Buffer {
	return lw.buffer
}

func (lw *LogWriter) GetWriteForPath(path string) (io.Writer, error) {
	// it will fail if the file already exists
	file, err := os.OpenFile(GetLogPath(path, lw._type), os.O_WRONLY|os.O_CREATE|os.O_EXCL, os.ModePerm)
	if err != nil {
		return nil, err
	}
	lw.file = file
	writers := []io.Writer{
		lw.file,
		&lw.buffer,
	}

	lw.writer = io.MultiWriter(writers...)
	return lw.writer, nil
}

func (lw *LogWriter) GetWriter() (io.Writer, error) {
	if lw.writer != nil {
		return lw.writer, nil
	}
	if lw.owner == nil {
		return nil, errors.New("log writer: no owner provided")
	}
	if lw.owner.path == "" {
		return nil, errors.New("log writer: invalid path of owner")
	}
	return lw.GetWriteForPath(lw.owner.path)
}

func (lw *LogWriter) Close() error {
	if lw.file != nil {
		return lw.file.Close()
	}
	return nil
}

// func IoTest() {
// 	file, err := os.OpenFile("./x.x", os.O_WRONLY|os.O_CREATE, os.ModePerm)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer file.Close()

// 	// Create a buffered writer from the file
// 	bufferedWriter := bufio.NewWriter(file)

// 	// Write bytes to buffer
// 	bytesWritten, err := bufferedWriter.Write(
// 		[]byte("welcome"),
// 	)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Printf("Bytes written: %d\n", bytesWritten)
// 	bufferedWriter.Flush()
// }
