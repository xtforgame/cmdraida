package t1

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	// "encoding/json"
	"github.com/xtforgame/cmdraida/crcore"
	"io/ioutil"
	"os"
	"strconv"
	// "strings"
)

type ReporterT1 struct {
	path            string
	taskUid         string
	reporterOptions *crcore.ReporterOptions
	stdoutWriter    crcore.Writer
	stderrWriter    crcore.Writer
	resultWriter    crcore.Writer
}

func NewReporterT1(taskUid string, options *crcore.ReporterOptions) crcore.Reporter {
	path := filepath.Join(options.BasePath, taskUid)
	reporter := &ReporterT1{
		path:            path,
		taskUid:         taskUid,
		reporterOptions: options,
	}
	os.MkdirAll(path, os.ModePerm)

	var err error

	if reporter.stdoutWriter, err = NewLogWriter("stdout", reporter); err != nil {
		return nil
	}
	if reporter.stderrWriter, err = NewLogWriter("stderr", reporter); err != nil {
		return nil
	}
	if reporter.resultWriter, err = NewLogWriter("result", reporter); err != nil {
		return nil
	}
	return reporter
}

func (reporter *ReporterT1) GetTaskUid() string {
	return reporter.taskUid
}

func (reporter *ReporterT1) GetReporterOptions() *crcore.ReporterOptions {
	return reporter.reporterOptions
}

func (reporter *ReporterT1) GetStdoutWriter() (crcore.Writer, error) {
	return reporter.stdoutWriter, nil
}

func (reporter *ReporterT1) GetStderrWriter() (crcore.Writer, error) {
	return reporter.stderrWriter, nil
}

func (reporter *ReporterT1) GetResultWriter() (crcore.Writer, error) {
	return reporter.resultWriter, nil
}

func (reporter *ReporterT1) ProduceResultLog(resultOutput crcore.Writer, finalStatus *crcore.FinalStatus) (*crcore.ResultLog, error) {
	resultLog := &crcore.ResultLog{}
	resultLog.TaskUid = reporter.GetTaskUid()
	if finalStatus.Error != nil {
		resultLog.Error = fmt.Sprintf("%s", finalStatus.Error)
	}

	resultLog.IsKilledByCommand = finalStatus.IsKilledByCommand
	resultLog.IsKilledByTimeout = finalStatus.IsKilledByTimeout
	resultLog.IsTerminatedByTimeout = finalStatus.IsTerminatedByTimeout

	resultLog.Command = &finalStatus.CommandType
	// outputBytes := finalStatus.Task.GetFullOutput()
	// resultLog.Output = string(outputBytes)
	// if resultLog.Command.ExpectJsonResult {
	// 	err := json.Unmarshal(outputBytes, &resultLog.JsonOutput)
	// 	if err == nil {
	// 		resultLog.Output = ""
	// 	}
	// }

	if finalStatus.Signaled {
		sigVal, _ := strconv.ParseInt(fmt.Sprintf("%d", finalStatus.Signal), 10, 64)
		resultLog.Signal = &crcore.SignalType{
			Name:  finalStatus.Signal.String(),
			Value: sigVal,
		}
	}

	if finalStatus.IsExitStatusValid {
		resultLog.ExitStatus = &crcore.ExitStatusType{
			Value: finalStatus.ExitStatus,
		}
	}
	b, _ := resultLog.ToJson()
	resultOutput.Write(b)
	return resultLog, nil
}

func (reporter *ReporterT1) ReadFinishedResultLog() (*crcore.ResultLog, error) {
	var resultLog *crcore.ResultLog
	bytes, err := ioutil.ReadFile(GetLogPath(filepath.Join(reporter.GetReporterOptions().BasePath, reporter.GetTaskUid()), "result"))
	if err == nil {
		resultLog = &crcore.ResultLog{}
		err = json.Unmarshal(bytes, resultLog)
		if err != nil {
			resultLog = nil
		}
	}
	return resultLog, err
}

func (reporter *ReporterT1) Close() error {
	if reporter.stdoutWriter != nil {
		reporter.stdoutWriter.Close()
		reporter.stdoutWriter = nil
	}

	if reporter.stderrWriter != nil {
		reporter.stderrWriter.Close()
		reporter.stderrWriter = nil
	}

	if reporter.resultWriter != nil {
		reporter.resultWriter.Close()
		reporter.resultWriter = nil
	}
	return nil
}
