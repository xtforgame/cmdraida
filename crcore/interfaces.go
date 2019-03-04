package crcore

import (
	"io"
)

type Writer interface {
	io.Writer
	Close() error
}

type Reporter interface {
	GetTaskUid() string
	GetReporterOptions() *ReporterOptions
	GetStdoutWriter() (Writer, error)
	GetStderrWriter() (Writer, error)
	GetResultWriter() (Writer, error)
	ProduceResultLog(resultOutput Writer, finalStatus *FinalStatus) (*ResultLog, error)
	ReadFinishedResultLog() (*ResultLog, error)
	Close() error
}

type TaskManager interface {
	GetBasePath() string
	CreateReporter(taskUid string, options *ReporterOptions) Reporter
}

type Task interface {
}
