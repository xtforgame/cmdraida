package crcore

import (
	"os/exec"
	"syscall"
)

func WaitFinalStatus(cmd *exec.Cmd, ch *chan FinalStatus) {
	err := cmd.Wait()
	finalStatus := FinalStatus{
		Error: err,
	}
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0
			finalStatus.IsExitStatusValid = ok

			// This works on both Unix and Windows. Although package
			// syscall is generally platform dependent, WaitStatus is
			// defined for both Unix and Windows and in both cases has
			// an ExitStatus() method with the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				finalStatus.WaitStatus = status
				finalStatus.IsWaitStatusValid = ok
				finalStatus.ExitStatus = status.ExitStatus()
				finalStatus.Signaled = status.Signaled()
				finalStatus.Signal = status.Signal()
			}
		}
	} else {
		status := cmd.ProcessState.Sys().(syscall.WaitStatus)
		finalStatus.WaitStatus = status
		finalStatus.ExitStatus = status.ExitStatus()
		finalStatus.IsWaitStatusValid = true
		finalStatus.IsExitStatusValid = true
		finalStatus.Signaled = status.Signaled()
		finalStatus.Signal = status.Signal()
	}
	*ch <- finalStatus
}
