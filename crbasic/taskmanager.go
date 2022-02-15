package crbasic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/xtforgame/cmdraida/crcore"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
	"time"
)

type TaskWithCallback struct {
	Task     *TaskBase
	Callback func(result interface{})
}

type TaskManagerBase struct {
	basePath        string
	taskMap         map[string]*TaskBase
	maxLogNumber    uint64
	cmdQueue        chan crcore.CommandWithCallback
	taskQueue       chan TaskWithCallback
	waitWorkerStop  chan bool
	ReporterCreator crcore.ReporterCreator
}

func (taskManager *TaskManagerBase) GetBasePath() string {
	return taskManager.basePath
}

func (taskManager *TaskManagerBase) CreateReporter(taskUid string, options *crcore.ReporterOptions) crcore.Reporter {
	return taskManager.ReporterCreator(taskUid, options)
}

func (taskManager *TaskManagerBase) createTask(commandType crcore.CommandType) *TaskBase {
	var task *TaskBase
	var err error
	var retry int
	for task == nil {
		taskManager.maxLogNumber++
		retry++
		if retry > 10 {
			return nil
		}
		task = NewTaskBase(taskManager, fmt.Sprintf("%08d", taskManager.maxLogNumber), commandType)

		_, err = task.Reporter.GetStdoutWriter()
		if err != nil {
			task = nil
		}
	}
	return task
}

func (taskManager *TaskManagerBase) RunTask(commandType crcore.CommandType) *TaskBase {
	task := taskManager.createTask(commandType)
	if task != nil {
		task.Run()
	}
	return task
}

func (taskManager *TaskManagerBase) startWorker() {
	taskManager.waitWorkerStop = make(chan bool)
	stopped := false
	go func() {
		// fmt.Println("worker started")
		for {
			select {
			case command := <-taskManager.cmdQueue:
				// fmt.Println("command :", command)
				if command.IsTerminalCmd {
					stopped = true
				}
				cancel := make(chan interface{}, 1)
				crcore.CancelableAsync(func() interface{} {
					task := taskManager.createTask(command.CommandType)
					command.OnTaskCreated(task)
					if task != nil {
						task.Run()
					}
					return task
				}, command.Callback, cancel)
			case taskData := <-taskManager.taskQueue:
				// fmt.Println("taskData :", taskData)
				if taskData.Task.command.IsTerminalCmd {
					stopped = true
				}
				cancel := make(chan interface{}, 1)
				crcore.CancelableAsync(func() interface{} {
					if taskData.Task != nil {
						taskData.Task.Run()
					}
					return taskData
				}, taskData.Callback, cancel)
			}
			if stopped {
				break
			}
		}
		// fmt.Println("worker finished")
		taskManager.waitWorkerStop <- true
	}()
}

func (taskManager *TaskManagerBase) finishWorker() {
	taskManager.cmdQueue <- crcore.CommandWithCallback{
		CommandType: crcore.CommandType{
			IsTerminalCmd: true,
		},
	}
	select {
	case <-time.After(time.Second * 5):
		// fmt.Println("worker finished by timeout")
	case <-taskManager.waitWorkerStop:
		// fmt.Println("worker finished gracefully")
	}
	if taskManager.waitWorkerStop != nil {
		close(taskManager.waitWorkerStop)
		// taskManager.waitWorkerStop = nil
	}
}

func (taskManager *TaskManagerBase) TaskMap() map[string]*TaskBase {
	return taskManager.taskMap
}

func NewTaskManager(basePath string, ReporterCreator crcore.ReporterCreator) *TaskManagerBase {
	return &TaskManagerBase{
		basePath:        basePath,
		taskMap:         map[string]*TaskBase{},
		cmdQueue:        make(chan crcore.CommandWithCallback, 3),
		taskQueue:       make(chan TaskWithCallback, 3),
		ReporterCreator: ReporterCreator,
	}
}

func (taskManager *TaskManagerBase) Init() {
	fileInfos, err := ioutil.ReadDir(taskManager.basePath)
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
				taskManager.taskMap[parts[1]] = NewFinishedTaskBase(taskManager, fmt.Sprintf("%08d", logNumber))
				if taskManager.maxLogNumber < logNumber {
					taskManager.maxLogNumber = logNumber
				}
			}
		}
		// fmt.Println("fi :", path)
	}
	taskManager.startWorker()
}

func (taskManager *TaskManagerBase) Close() {
	for _, logData := range taskManager.taskMap {
		logData.Close()
	}
	taskManager.finishWorker()
	// if taskManager.cmdQueue != nil {
	// 	close(taskManager.cmdQueue)
	// 	taskManager.cmdQueue = nil
	// }
	// if taskManager.taskQueue != nil {
	// 	close(taskManager.taskQueue)
	// 	taskManager.taskQueue = nil
	// }
}

func (taskManager *TaskManagerBase) TestNewTask() *TaskBase {
	var task *TaskBase
	var err error
	var retry int
	for task == nil {
		taskManager.maxLogNumber++
		retry++
		if retry > 10 {
			return nil
		}
		task = NewTaskBase(taskManager, fmt.Sprintf("%08d", taskManager.maxLogNumber), crcore.CommandType{
			Command: "bash",
			Args:    []string{"-c", "echo xxx;sleep 2;echo ooo"},
			Timeouts: crcore.TimeoutsType{
				Proccess:    1000,
				AfterKilled: 1500,
			},
		})
		// task = NewTaskBase(taskManager, taskManager.maxLogNumber, taskManager.basePath, CommandType{
		// 	Command: "bash",
		// 	Args: []string{"-c", "echo xxx;sleep 2;echo ooo"},
		// 	Timeouts: TimeoutsType{
		// 		Proccess: 1000,
		// 		AfterKilled: 500,
		// 	},
		// })
		// task = NewTaskBase(taskManager, taskManager.maxLogNumber, taskManager.basePath, CommandType{
		// 	Command: "azpbrctl",
		// 	Args: []string{"-h"},
		// })
		// task = NewTaskBase(taskManager, taskManager.maxLogNumber, taskManager.basePath, CommandType{
		// 	Command: "restic",
		// 	Args: []string{"-h"},
		// })

		// task = NewTaskBase(taskManager, taskManager.maxLogNumber, taskManager.basePath, CommandType{
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

		// task = NewTaskBase(taskManager, taskManager.maxLogNumber, taskManager.basePath, CommandType{
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

func (taskManager *TaskManagerBase) AddCommand(command crcore.CommandType, onTaskCreated func(*TaskBase), callback func(interface{})) {
	taskManager.cmdQueue <- crcore.CommandWithCallback{
		CommandType: command,
		OnTaskCreated: func(task interface{}) {
			t, ok := task.(*TaskBase)
			if ok {
				onTaskCreated(t)
			} else {
				onTaskCreated(nil)
			}
		},
		Callback: callback,
	}
}

func (taskManager *TaskManagerBase) AddTask(command crcore.CommandType, callback func(interface{})) *TaskBase {
	task := taskManager.createTask(command)
	if task != nil {
		taskManager.taskQueue <- TaskWithCallback{
			Task:     task,
			Callback: callback,
		}
	}
	return task
}

func (taskManager *TaskManagerBase) TestNewTask2() *TaskBase {
	task := taskManager.AddTask(crcore.CommandType{
		Command: "bash",
		Args:    []string{"-c", "echo xxx;sleep 1;echo ooo"},
		Timeouts: crcore.TimeoutsType{
			Proccess:    3000,
			AfterKilled: 1500,
		},
	}, func(r interface{}) {

	})
	count := 0
	for {
		result := task.ResultLog()
		if result != nil {
			// fmt.Println("result :", result)
			break
		}
		count++
		if count > 5 {
			// fmt.Println("count > 5 :", count)
			break
		}
		time.Sleep(1 * time.Second)
	}
	// fmt.Println("count :", count)
	return task
}

func ResultLogsToJson(resultLogs []*crcore.ResultLog) ([]byte, error) {
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

func (taskManager *TaskManagerBase) GetTaskListJson() ([]byte, error) {
	taskMap := taskManager.TaskMap()
	tasks := TaskSlice{}
	for _, v := range taskMap {
		tasks = append(tasks, v)
	}
	sort.Sort(tasks)

	results := []*crcore.ResultLog{}
	for _, task := range tasks {
		results = append(results, task.ResultLog())

	}
	return ResultLogsToJson(results)
}
