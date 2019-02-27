package t1

import (
	"path/filepath"
)

func GetLogPath(path, _type string) string {
	return filepath.Join(path, _type+".log")
}
