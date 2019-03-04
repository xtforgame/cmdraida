package t1

import (
	"bytes"
	"github.com/xtforgame/cmdraida/crcore"
	"io"
	"os"
	"path/filepath"
)

type LogWriter struct {
	_type  string
	owner  crcore.Reporter
	file   *os.File
	buffer bytes.Buffer
	writer io.Writer
}

func NewLogWriter(_type string, owner crcore.Reporter) (*LogWriter, error) {
	lw := &LogWriter{
		_type:  _type,
		owner:  owner,
		buffer: bytes.Buffer{},
	}

	path := GetLogPath(filepath.Join(owner.GetReporterOptions().BasePath, owner.GetTaskUid()), lw._type)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, os.ModePerm)
	if err != nil {
		return nil, err
	}
	lw.file = file
	writers := []io.Writer{
		lw.file,
		&lw.buffer,
	}

	lw.writer = io.MultiWriter(writers...)
	return lw, nil
}

func (lw *LogWriter) Buffer() bytes.Buffer {
	return lw.buffer
}

func (lw *LogWriter) Write(p []byte) (n int, err error) {
	return lw.writer.Write(p)
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
