package log

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

const logBaseDir = "./_logs"

type LogType int

const (
	Auto LogType = iota
	Manual
)

const (
	LOG_FSIZE_LIMIT   = 1_000_000
	FNAME_TIME_FORMAT = "20060102150405"
)

var (
	logTypes          = []string{"auto", "manual"}
	logLastLabels     = make([]string, len(logTypes))
	logLastLabelsCols = make([]int, len(logTypes))
	logBuffers        = make([]map[string][]string, len(logTypes))

	validLabelChars = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ_")
)

func Setup() error {
	for _, logType := range logTypes {
		path := fmt.Sprintf("%s/%s", logBaseDir, logType)
		err := os.MkdirAll(path, 0760)
		if err != nil {
			return errors.Wrapf(err, "failed to mkdir for log type: %s", logType)
		}
	}

	for i := range logTypes {
		logLastLabels[i] = ""
		logBuffers[i] = map[string][]string{}
	}

	return nil
}
