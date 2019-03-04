package crcore

import (
	"syscall"
)

type TimeoutsType struct {
	Proccess    uint64 `json:"proccess,"`
	AfterKilled uint64 `json:"after-killed,"`
}

type CommandType struct {
	Command          string       `json:"command,"`
	Args             []string     `json:"args,"`
	ExpectJsonResult bool         `json:"-"`
	Timeouts         TimeoutsType `json:"timeouts,"`
	Env              []string     `json:"env,"`
	Dir              string       `json:"dir,"`
	IsTerminalCmd    bool
}

type SignalType struct {
	Value int64  `json:"value,"`
	Name  string `json:"message,"`
}

type ExitStatusType struct {
	Value int `json:"value,"`
}

type FinalStatus struct {
	CommandType
	Task                  Task
	Error                 error
	WaitStatus            syscall.WaitStatus
	IsWaitStatusValid     bool
	ExitStatus            int
	IsExitStatusValid     bool
	Signaled              bool
	Signal                syscall.Signal
	IsKilledByCommand     bool
	IsKilledByTimeout     bool
	IsTerminatedByTimeout bool
}

// ================

type CommandWithCallback struct {
	CommandType
	Callback func(result interface{})
}

type ReporterOptions struct {
	BasePath string
}

type ReporterCreator func(taskUid string, options *ReporterOptions) Reporter
