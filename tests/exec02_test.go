package tests

import (
	"testing"
	// "errors"
	// "bufio"
	// "github.com/xtforgame/cmdraida"
	"github.com/xtforgame/cmdraida/crcore"
	// "io"
	"os"
)

var exec02TestFolder = "../tmp/test/exec02"
var exec02TestFile = exec02TestFolder + "/x.x"

func TestExec02Run01(t *testing.T) {
	os.RemoveAll(exec02TestFolder)
	os.MkdirAll(exec02TestFolder, os.ModePerm)

	crcore.NewTaskManager(exec02TestFolder)
}
