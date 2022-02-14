package tests

import (
	"testing"
	// "errors"
	// "bufio"
	// "github.com/xtforgame/cmdraida"
	// "github.com/xtforgame/cmdraida/crcore"
	"github.com/xtforgame/cmdraida/crbasic"
	"github.com/xtforgame/cmdraida/t1"
	// "io"
	"os"
)

var exec02TestFolder = "../tmp/test/exec02"

func TestExec02Run01(t *testing.T) {
	os.RemoveAll(exec02TestFolder)
	os.MkdirAll(exec02TestFolder, os.ModePerm)

	manager := crbasic.NewTaskManager(exec02TestFolder, t1.NewReporterT1)
	manager.Init()
	manager.TestNewTask()
	manager.TestNewTask2()

	defer manager.Close()
}
