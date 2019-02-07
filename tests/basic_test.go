package tests

import (
	"testing"
	// "errors"
	// "bufio"
	"io"
	"os"
)

// func TestSomething(t *testing.T) {
// 	t.Skip()
// }

// func TestAdd(t *testing.T) {
// 	if 1+2 == 3 {
// 		t.Log("mymath.Add PASS")
// 	} else {
// 		t.Error("mymath.Add FAIL")
// 	}
// }

// func TestIo(t *testing.T) {
// 	os.RemoveAll("../tmp/test")
// 	os.MkdirAll("../tmp/test", os.ModePerm)
// 	file, err := os.OpenFile(testFile, os.O_WRONLY|os.O_CREATE, os.ModePerm)
// 	if err != nil {
// 		t.Error("open file FAIL")
// 	}
// 	defer file.Close()

// 	// Create a buffered writer from the file
// 	bufferedWriter := bufio.NewWriter(file)

// 	// Write bytes to buffer
// 	bytesWritten, err := bufferedWriter.Write(
// 		[]byte("welcome"),
// 	)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	t.Log("Bytes written: ", bytesWritten)
// 	// bufferedWriter.Flush()
// }

var basicTestFolder = "../tmp/test/basic"
var basicTestFile = basicTestFolder + "/x.x"

func TestWrite(t *testing.T) {
	os.RemoveAll(basicTestFolder)
	os.MkdirAll(basicTestFolder, os.ModePerm)
	// it will fail if the file already exists
	file, err := os.OpenFile(basicTestFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, os.ModePerm)
	if err != nil {
		t.Error("open file FAIL")
	}

	fw := &FileWriter{
		file: file,
	}
	writers := []io.Writer{
		file,
		fw,
	}

	writer := io.MultiWriter(writers...)
	writer.Write([]byte("XXXXXXXX"))
	fw.Close()
	writer.Write([]byte("OOOOOO"))
}
