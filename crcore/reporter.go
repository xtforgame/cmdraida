package crcore

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	// "strconv"
	// "strings"
)

type Reporter struct {
	number       uint64
	path         string
	logWriters   map[string]*LogWriter
	stdoutWriter io.Writer
	stderrWriter io.Writer
	resultWriter io.Writer
}

func NewReporter(number uint64, basePath string) Reporter {
	path := filepath.Join(basePath, "log_"+fmt.Sprintf("%08d", number))
	return Reporter{
		number:     number,
		path:       path,
		logWriters: map[string]*LogWriter{},
	}
}

func (reporter *Reporter) getWriter(_type string) (io.Writer, error) {
	if _, found := reporter.logWriters[_type]; !found {
		reporter.logWriters[_type] = NewLogWriter(_type, reporter)
	}
	writer, err := reporter.logWriters[_type].GetWriter()
	if err != nil {
		reporter.logWriters[_type] = nil
	}
	return writer, err
}

func (reporter *Reporter) GetStdoutWriter() (io.Writer, error) {
	if reporter.stdoutWriter != nil {
		return reporter.stdoutWriter, nil
	}
	if err := os.MkdirAll(reporter.path, os.ModePerm); err != nil {
		return nil, err
	}
	stdoutWriter, err := reporter.getWriter("stdout")
	if err != nil {
		return nil, err
	}
	fullWriter, err := reporter.getWriter("full")
	if err != nil {
		return nil, err
	}
	writers := []io.Writer{
		stdoutWriter,
		fullWriter,
	}

	fileAndStdoutWriter := io.MultiWriter(writers...)
	reporter.stdoutWriter = fileAndStdoutWriter
	return reporter.stdoutWriter, nil
}

func (reporter *Reporter) GetStderrWriter() (io.Writer, error) {
	if reporter.stderrWriter != nil {
		return reporter.stderrWriter, nil
	}
	if err := os.MkdirAll(reporter.path, os.ModePerm); err != nil {
		return nil, err
	}
	stderrWriter, err := reporter.getWriter("stderr")
	if err != nil {
		return nil, err
	}
	fullWriter, err := reporter.getWriter("full")
	if err != nil {
		return nil, err
	}
	writers := []io.Writer{
		stderrWriter,
		fullWriter,
	}

	fileAndStdoutWriter := io.MultiWriter(writers...)
	reporter.stderrWriter = fileAndStdoutWriter
	return reporter.stderrWriter, nil
}

func (reporter *Reporter) GetResultWriter() (io.Writer, error) {
	if reporter.resultWriter != nil {
		return reporter.resultWriter, nil
	}
	if err := os.MkdirAll(reporter.path, os.ModePerm); err != nil {
		return nil, err
	}
	resultWriter, err := reporter.getWriter("result")
	if err != nil {
		return nil, err
	}
	writers := []io.Writer{
		resultWriter,
	}

	fileAndStdoutWriter := io.MultiWriter(writers...)
	reporter.resultWriter = fileAndStdoutWriter
	return reporter.resultWriter, nil
}

func (reporter *Reporter) GetFullOutput() []byte {
	lw, found := reporter.logWriters["full"]
	if !found {
		return []byte("")
	}
	b := lw.Buffer()
	return b.Bytes()
}

func (reporter *Reporter) Close() {
	for _, logWriter := range reporter.logWriters {
		logWriter.Close()
	}
	reporter.stdoutWriter = nil
	reporter.stderrWriter = nil
}
