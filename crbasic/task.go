package crbasic

import (
	"errors"
	// "fmt"
	"io"
	"os/exec"
	// "path/filepath"
	"github.com/xtforgame/cmdraida/crcore"
	// "syscall"
	"time"
)

type TaskBase struct {
	Reporter      crcore.Reporter
	TaskUid       string
	command       *crcore.CommandType
	manager       crcore.TaskManager
	cmd           *exec.Cmd
	cmdChan       chan crcore.FinalStatus
	terminateChan chan string
	resultLog     *crcore.ResultLog
}

func (task *TaskBase) ResultLog() *crcore.ResultLog { return task.resultLog }

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

func NewFinishedTaskBase(manager crcore.TaskManager, taskUid string) *TaskBase {
	task := &TaskBase{
		manager: manager,
		Reporter: manager.CreateReporter(taskUid, &crcore.ReporterOptions{
			BasePath: manager.GetBasePath(),
		}),
	}
	task.resultLog, _ = task.Reporter.ReadFinishedResultLog()
	return task
}

func NewTaskBase(manager crcore.TaskManager, taskUid string, commandType crcore.CommandType) *TaskBase {
	if commandType.Timeouts.Proccess == 0 {
		commandType.Timeouts.Proccess = 24 * 60 * 60 * 1000
	}

	if commandType.Timeouts.AfterKilled == 0 {
		commandType.Timeouts.AfterKilled = 1000
	}

	return &TaskBase{
		manager: manager,
		Reporter: manager.CreateReporter(taskUid, &crcore.ReporterOptions{
			BasePath: manager.GetBasePath(),
		}),
		command: &commandType,
	}
}

func (taskA *TaskBase) IsLessThan(taskB *TaskBase) bool {
	return taskA.Reporter.GetTaskUid() < taskB.Reporter.GetTaskUid()
}

// https://github.com/golang/go/issues/7938
func (task *TaskBase) Exec(stdout, stderr, resultOutput crcore.Writer) *crcore.FinalStatus {
	var fStatus = crcore.FinalStatus{}
	// cmd := exec.Command("./azpbrctl")
	if task.command.Args != nil {
		task.cmd = exec.Command(task.command.Command, task.command.Args...)
	} else {
		task.cmd = exec.Command(task.command.Command)
	}
	if task.command.Env != nil {
		task.cmd.Env = append(task.cmd.Env, task.command.Env...)
	}
	if task.command.Dir != "" {
		task.cmd.Dir = task.command.Dir
	}
	task.cmd.Stdout = stdout
	task.cmd.Stderr = stderr
	task.cmdChan = make(chan crcore.FinalStatus)
	task.terminateChan = make(chan string)
	startErr := task.cmd.Start()
	if startErr == nil {
		go crcore.WaitFinalStatus(task.cmd, &task.cmdChan)

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
		// if exiterr, ok := err.(*exec.ExitError); ok {
		// 	// The program has exited with an exit code != 0

		// 	// This works on both Unix and Windows. Although package
		// 	// syscall is generally platfotaskManager dependent, WaitStatus is
		// 	// defined for both Unix and Windows and in both cases has
		// 	// an ExitStatus() method with the same signature.
		// 	// if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
		// 	// 	fmt.Printf("Exit Status: %d\n", status.ExitStatus())
		// 	// }
		// } else {
		// 	// fmt.Printf("task.cmd.Wait: %v\n", err)
		// }
		done <- err
	}()

	time.Sleep(time.Second)
	if task.cmd.Process != nil {
		task.cmd.Process.Kill()
		err = <-done
	}

	return err
}

func (task *TaskBase) Run() (*crcore.FinalStatus, error) {
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
