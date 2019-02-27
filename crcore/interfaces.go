package crcore

import (
	"io"
)

type Writer interface {
	io.Writer
	Close() error
}

type Reporter interface {
	GetNumber() uint64
	GetPath() string
	GetStdoutWriter() (Writer, error)
	GetStderrWriter() (Writer, error)
	GetResultWriter() (Writer, error)
	ProduceResultLog(resultOutput Writer, finalStatus *FinalStatus) (*ResultLog, error)
	ReadFinishedResultLog() (*ResultLog, error)
	Close() error
}
