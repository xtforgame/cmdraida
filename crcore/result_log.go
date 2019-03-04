package crcore

import (
	"bytes"
	"encoding/json"
)

type ResultLog struct {
	TaskUid               string          `json:"task,"`
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
