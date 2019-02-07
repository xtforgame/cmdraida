package tests

import (
	"bytes"
	"fmt"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func WaitCmd(cmd *exec.Cmd, ch *chan error, t *testing.T) {
	err := cmd.Wait()
	if exiterr, ok := err.(*exec.ExitError); ok {
		// The program has exited with an exit code != 0

		// This works on both Unix and Windows. Although package
		// syscall is generally platform dependent, WaitStatus is
		// defined for both Unix and Windows and in both cases has
		// an ExitStatus() method with the same signature.
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			t.Logf("Exit Status: %d\n", status.ExitStatus())
		}
	} else {
		t.Logf("cmd.Wait: %v\n", err)
	}
	*ch <- err
}

func TestExec01Run01(t *testing.T) {
	// cmd := exec.Command("./azpbrctl")
	cmd := exec.Command("bash", "-c", "echo xxx;sleep 2;echo ooo")
	var err error
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	ch := make(chan error)
	cmd.Start()
	go WaitCmd(cmd, &ch, t)

	time.Sleep(time.Second)
	if cmd.Process != nil {
		cmd.Process.Kill()
		err = <-ch
	}

	if err == nil {
		t.Fatal("no killed error")
	}

	// fmt.Println("status:", err)
	if outb.String() != "xxx\n" {
		t.Fatal("wrong output", outb.String())
	}
	if errb.String() != "" {
		t.Fatal("wrong output", errb.String())
	}
}

func TestExec01Run02(t *testing.T) {
	// cmd := exec.Command("./azpbrctl")
	cmd := exec.Command("bash", "-c", "echo xxx;sleep 2;echo ooo")
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	ch := make(chan error)
	cmd.Start()
	go WaitCmd(cmd, &ch, t)
	select {
	case <-time.After(1 * time.Second):
		if cmd.Process != nil {
			if err := cmd.Process.Kill(); err != nil {
				fmt.Println("failed to kill process: ", err)
			}
		}
		fmt.Println("process killed as timeout reached")
	case err := <-ch:
		if err != nil {
			t.Logf("process finished with error = %v", err)
		}
		fmt.Println("process finished successfully")
	}

	// fmt.Println("status:", err)
	if outb.String() != "xxx\n" {
		t.Fatal("wrong output", outb.String())
	}
	if errb.String() != "" {
		t.Fatal("wrong output", errb.String())
	}
}
