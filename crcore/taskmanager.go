package crcore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
	"time"
)

type CommandWithCallback struct {
	CommandType
	callback func(result interface{})
}

type ReporterCreator func(number uint64, basePath string) Reporter

type TaskManager struct {
	rootFolder     string
	taskMap        map[string]*TaskBase
	maxLogNumber   uint64
	cmdQueue       chan CommandWithCallback
	waitWorkerStop chan bool
	CreateReporter ReporterCreator
}

func (taskManager *TaskManager) runTask(commandType CommandType) *TaskBase {
	var task *TaskBase
	var err error
	var retry int
	for task == nil {
		taskManager.maxLogNumber++
		retry++
		if retry > 10 {
			return nil
		}
		task = NewTaskBase(taskManager, taskManager.maxLogNumber, taskManager.rootFolder, commandType)

		_, err = task.Reporter.GetStdoutWriter()
		if err != nil {
			task = nil
		}
	}
	task.Run()
	return task
}

func (taskManager *TaskManager) startWorker() {
	taskManager.waitWorkerStop = make(chan bool)
	go func() {
		fmt.Println("worker started")
		for command := range taskManager.cmdQueue {
			fmt.Println("command :", command)
			if command.isTerminalCmd {
				break
			}
			cancel := make(chan interface{}, 1)
			CancelableAsync(func() interface{} {
				return taskManager.runTask(command.CommandType)
			}, command.callback, cancel)
		}
		fmt.Println("worker finished")
		taskManager.waitWorkerStop <- true
	}()
}

func (taskManager *TaskManager) finishWorker() {
	taskManager.cmdQueue <- CommandWithCallback{
		CommandType: CommandType{
			isTerminalCmd: true,
		},
	}
	select {
	case <-time.After(time.Second * 5):
		fmt.Println("worker finished by timeout")
	case <-taskManager.waitWorkerStop:
		fmt.Println("worker finished gracefully")
	}
	if taskManager.waitWorkerStop != nil {
		close(taskManager.waitWorkerStop)
		// taskManager.waitWorkerStop = nil
	}
}

func (taskManager *TaskManager) TaskMap() map[string]*TaskBase {
	return taskManager.taskMap
}

func NewTaskManager(rootFolder string, createReporter ReporterCreator) *TaskManager {
	return &TaskManager{
		rootFolder:     rootFolder,
		taskMap:        map[string]*TaskBase{},
		cmdQueue:       make(chan CommandWithCallback, 3),
		CreateReporter: createReporter,
	}
}

func (taskManager *TaskManager) Init() {
	fileInfos, err := ioutil.ReadDir(taskManager.rootFolder)
	if err != nil {
		panic(err)
	}
	for _, fi := range fileInfos {
		name := fi.Name()
		// fmt.Println("fi :", name)
		parts := strings.Split(name, "_")
		if len(parts) == 2 && parts[0] == "log" {
			logNumber, err := strconv.ParseUint(parts[1], 10, 64)
			if err == nil {
				taskManager.taskMap[parts[1]] = NewFinishedTaskBase(taskManager, logNumber, taskManager.rootFolder)
				if taskManager.maxLogNumber < logNumber {
					taskManager.maxLogNumber = logNumber
				}
			}
		}
		// fmt.Println("fi :", path)
	}
	taskManager.startWorker()
}

func (taskManager *TaskManager) Close() {
	for _, logData := range taskManager.taskMap {
		logData.Close()
	}
	taskManager.finishWorker()
	// if taskManager.cmdQueue != nil {
	// 	close(taskManager.cmdQueue)
	// 	taskManager.cmdQueue = nil
	// }
}

func (taskManager *TaskManager) TestNewTask() *TaskBase {
	var task *TaskBase
	var err error
	var retry int
	for task == nil {
		taskManager.maxLogNumber++
		retry++
		if retry > 10 {
			return nil
		}
		task = NewTaskBase(taskManager, taskManager.maxLogNumber, taskManager.rootFolder, CommandType{
			Command: "bash",
			Args:    []string{"-c", "echo xxx;sleep 2;echo ooo"},
			Timeouts: TimeoutsType{
				Proccess:    1000,
				AfterKilled: 1500,
			},
		})
		// task = NewTaskBase(taskManager, taskManager.maxLogNumber, taskManager.rootFolder, CommandType{
		// 	Command: "bash",
		// 	Args: []string{"-c", "echo xxx;sleep 2;echo ooo"},
		// 	Timeouts: TimeoutsType{
		// 		Proccess: 1000,
		// 		AfterKilled: 500,
		// 	},
		// })
		// task = NewTaskBase(taskManager, taskManager.maxLogNumber, taskManager.rootFolder, CommandType{
		// 	Command: "azpbrctl",
		// 	Args: []string{"-h"},
		// })
		// task = NewTaskBase(taskManager, taskManager.maxLogNumber, taskManager.rootFolder, CommandType{
		// 	Command: "restic",
		// 	Args: []string{"-h"},
		// })

		// task = NewTaskBase(taskManager, taskManager.maxLogNumber, taskManager.rootFolder, CommandType{
		// 	Command: taskManager.cliHelper.CmdAzprbctl(),
		// 	Args: []string{
		// 		"-m",
		// 		"-stanza",
		// 		"azpbr",
		// 		// "-o",
		// 		// filepath.Join(taskManager.cliHelper.WebDataPath(), "./report.json"),
		// 		taskManager.cliHelper.AzpbrSrcPath(),
		// 		taskManager.cliHelper.AzpbrDistPath(),
		// 	},
		// })

		// task = NewTaskBase(taskManager, taskManager.maxLogNumber, taskManager.rootFolder, CommandType{
		// 	Command: taskManager.cliHelper.CmdRestic(),
		// 	Args: []string{
		// 		"restore",
		// 		"latest",
		// 		"--target",
		// 		taskManager.cliHelper.AzpbrDistPath(),
		// 	},
		// })

		_, err = task.Reporter.GetStdoutWriter()
		if err != nil {
			task = nil
		}
	}
	task.Run()
	return task
}

func (taskManager *TaskManager) AddTask(command CommandType, callback func(interface{})) {
	taskManager.cmdQueue <- CommandWithCallback{
		CommandType: command,
		callback:    callback,
	}
}

func ResultLogsToJson(resultLogs []*ResultLog) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(resultLogs); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
	// return json.Marshal(resultLogs)
}

func (taskManager *TaskManager) GetTaskListJson() ([]byte, error) {
	taskMap := taskManager.TaskMap()
	tasks := TaskSlice{}
	for _, v := range taskMap {
		tasks = append(tasks, v)
	}
	sort.Sort(tasks)

	results := []*ResultLog{}
	for _, task := range tasks {
		results = append(results, task.ResultLog())

	}
	return ResultLogsToJson(results)
}
