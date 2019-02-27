package crcore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	// "path/filepath"
	"syscall"
	"time"
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
	isTerminalCmd    bool
}

type SignalType struct {
	Value int64  `json:"value,"`
	Name  string `json:"message,"`
}

type ExitStatusType struct {
	Value int `json:"value,"`
}

type ResultLog struct {
	TaskNumber            uint64          `json:"task,"`
	Command               *CommandType    `json:"command,"`
	Error                 string          `json:"error,"`
	Signal                *SignalType     `json:"signal,"`
	ExitStatus            *ExitStatusType `json:"exit-status,"`
	IsKilledByCommand     bool            `json:"killed-by-command,"`
	IsKilledByTimeout     bool            `json:"killed-by-timeout,"`
	IsTerminatedByTimeout bool            `json:"terminated-by-timeout,"`
	Output                string          `json:"output,"`
	JsonOutput            json.RawMessage `json:"json-output,omitempty"`
	Stdout                string          `json:"stdout,"`
	JsonStdout            json.RawMessage `json:"json-stdout,omitempty"`
	Stderr                string          `json:"stderr,"`
	JsonStderr            json.RawMessage `json:"json-stderr,omitempty"`
}

func (resultLog *ResultLog) ToJson() ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(resultLog); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type FinalStatus struct {
	CommandType
	Task                  *TaskBase
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

type TaskBase struct {
	Reporter      Reporter
	command       *CommandType
	manager       *TaskManager
	cmd           *exec.Cmd
	cmdChan       chan FinalStatus
	terminateChan chan string
	resultLog     *ResultLog
}

func (task *TaskBase) ResultLog() *ResultLog { return task.resultLog }

type TaskSlice []*TaskBase

func (t TaskSlice) Len() int {
	return len(t)
}

func (t TaskSlice) Less(i, j int) bool {
	return t[i].IsLessThan(t[j])
}

func (t TaskSlice) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func NewFinishedTaskBase(manager *TaskManager, number uint64, basePath string) *TaskBase {
	task := &TaskBase{
		manager:  manager,
		Reporter: manager.CreateReporter(number, basePath),
	}
	task.resultLog, _ = task.Reporter.ReadFinishedResultLog()
	return task
}

func NewTaskBase(manager *TaskManager, number uint64, basePath string, commandType CommandType) *TaskBase {
	if commandType.Timeouts.Proccess == 0 {
		commandType.Timeouts.Proccess = 24 * 60 * 60 * 1000
	}

	if commandType.Timeouts.AfterKilled == 0 {
		commandType.Timeouts.AfterKilled = 1000
	}

	return &TaskBase{
		manager:  manager,
		Reporter: manager.CreateReporter(number, basePath),
		command:  &commandType,
	}
}

func (taskA *TaskBase) IsLessThan(taskB *TaskBase) bool {
	return taskA.Reporter.GetNumber() < taskB.Reporter.GetNumber()
}

// https://github.com/golang/go/issues/7938
func (task *TaskBase) Exec(stdout, stderr, resultOutput Writer) *FinalStatus {
	var fStatus = FinalStatus{}
	// cmd := exec.Command("./azpbrctl")
	if task.command.Args != nil {
		task.cmd = exec.Command(task.command.Command, task.command.Args...)
	} else {
		task.cmd = exec.Command(task.command.Command)
	}
	task.cmd.Stdout = stdout
	task.cmd.Stderr = stderr
	task.cmdChan = make(chan FinalStatus)
	task.terminateChan = make(chan string)
	startErr := task.cmd.Start()
	if startErr == nil {
		go WaitFinalStatus(task.cmd, &task.cmdChan)

		select {
		case reason := <-task.terminateChan:
			{
				if task.cmd.Process != nil {
					task.cmd.Process.Kill()
				}
				select {
				case <-time.After(time.Millisecond * time.Duration(task.command.Timeouts.AfterKilled)):
					fStatus.Error = errors.New(reason)
					fStatus.IsTerminatedByTimeout = true
				case fStatus = <-task.cmdChan:
				}
				fStatus.IsKilledByCommand = true
			}
		case <-time.After(time.Millisecond * time.Duration(task.command.Timeouts.Proccess)):
			{
				if task.cmd.Process != nil {
					task.cmd.Process.Kill()
				}
				select {
				case <-time.After(time.Millisecond * time.Duration(task.command.Timeouts.AfterKilled)):
					fStatus.Error = errors.New("task: terminated due to timeout")
					fStatus.IsTerminatedByTimeout = true
				case fStatus = <-task.cmdChan:
				}
				fStatus.IsKilledByTimeout = true
			}
		case fStatus = <-task.cmdChan:
		}
	} else {
		fStatus.Error = startErr
	}

	fStatus.CommandType = *task.command
	fStatus.Task = task
	task.resultLog, _ = task.Reporter.ProduceResultLog(resultOutput, &fStatus)
	return &fStatus
}

func (task *TaskBase) ExecOld(command string, args []string, stdout, stderr io.Writer) error {
	var err error
	// cmd := exec.Command("./azpbrctl")
	task.cmd = exec.Command(command, args...)
	task.cmd.Stdout = stdout
	task.cmd.Stderr = stderr
	done := make(chan error)
	task.cmd.Start()
	go func() {
		err := task.cmd.Wait()
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0

			// This works on both Unix and Windows. Although package
			// syscall is generally platfotaskManager dependent, WaitStatus is
			// defined for both Unix and Windows and in both cases has
			// an ExitStatus() method with the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				fmt.Printf("Exit Status: %d\n", status.ExitStatus())
			}
		} else {
			fmt.Printf("task.cmd.Wait: %v\n", err)
		}
		done <- err
	}()

	time.Sleep(time.Second)
	if task.cmd.Process != nil {
		task.cmd.Process.Kill()
		err = <-done
	}

	return err
}

func (task *TaskBase) Run() (*FinalStatus, error) {
	outWriter, err := task.Reporter.GetStdoutWriter()
	if err != nil {
		return nil, err
	}
	errWriter, err := task.Reporter.GetStderrWriter()
	if err != nil {
		return nil, err
	}
	resultWriter, err := task.Reporter.GetResultWriter()
	if err != nil {
		return nil, err
	}
	return task.Exec(outWriter, errWriter, resultWriter), nil
}

func (task *TaskBase) Kill() {
	go func() {
		task.terminateChan <- "killed by command"
	}()
}

func (task *TaskBase) Close() {
	task.Reporter.Close()
	// if task.cmdChan != nil {
	// 	close(task.cmdChan)
	// 	task.cmdChan = nil
	// }
	// if task.terminateChan != nil {
	// 	close(task.terminateChan)
	// 	task.terminateChan = nil
	// }
}
