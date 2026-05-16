package log

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pericles-tpt/ownapi/utility"
	"github.com/pkg/errors"
)

func WriteLogs(t LogType, label string, columnNames []string, values [][]any, flush bool) error {
	if len(values) == 0 {
		return errors.New("no rows provided")
	}
	if label == "" {
		return errors.New("invalid `label` provided, must be non-empty")
	}
	_, _, missingVals := utility.ContainsAll([]rune(label), validLabelChars)
	if len(missingVals) > 0 {
		return errors.Errorf("invalid `label` provided, contains invalid chars (not uppercase letters or underscores) at indices: %v", missingVals)
	}
	firstRowLen := len(values[0])
	if label != logLastLabels[t] {
		logLastLabels[t] = label
		logLastLabelsCols[t] = len(columnNames)
	}
	if firstRowLen != logLastLabelsCols[t] {
		return errors.Errorf("invalid number of vals provided in first row for log label %s, exp: %d, got: %d", label, logLastLabelsCols[t], firstRowLen)
	}

	rowLines := make([]string, 0, len(values))
	rowLens := make([]int, 0, len(values))
	allRowsEqualLen := true
	for _, r := range values {
		if len(r) != 0 {
			newRow := make([]any, 0, len(r)+1)
			newRow = append(newRow, time.Now().Format(FNAME_TIME_FORMAT))
			newRow = append(newRow, r...)
			rowLines = append(rowLines, strings.Join(utility.AnyToString(newRow), ","))
		}
		allRowsEqualLen = allRowsEqualLen && len(r) == firstRowLen
		rowLens = append(rowLens, len(r))
	}
	allZero := len(rowLines) == 0
	if allZero {
		return errors.Errorf("all rows are empty: %v", values)
	}
	if !allRowsEqualLen {
		return errors.Errorf("one or more rows don't match the len of the first row, row lens: %v", rowLens)
	}

	var (
		buf []string
		ok  bool
	)
	if buf, ok = logBuffers[t][label]; !ok {
		logBuffers[t][label] = make([]string, 0, 1)
	}
	logBuffers[t][label] = append(buf, rowLines...)

	path, new, err := getLogPathToWrite(t, label, getSizeOfBuffer(logBuffers[t][label]))
	if err != nil {
		return errors.Wrap(err, "failed to get log path for writing")
	}

	appendLines := make([]string, 0, 2)
	if new {
		header := make([]string, 0, len(values)+1)
		header = append(header, "TIMESTAMP")
		header = append(header, columnNames...)
		appendLines = append(appendLines, strings.Join(header, ","))
	}
	appendLines = append(appendLines, logBuffers[t][label]...)

	if flush {
		fileNewOrExisting := "existing"
		if !new {
			fileNewOrExisting = "new"
		}

		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0760)
		if err != nil {
			return errors.Wrapf(err, "failed to open %s file for appending: %s", fileNewOrExisting, path)
		}
		defer f.Close()

		content := fmt.Sprintf("%s\n", strings.Join(appendLines, "\n"))
		_, err = f.Write([]byte(content))
		if err != nil {
			return errors.Wrapf(err, "failed to write to %s file: %s", fileNewOrExisting, path)
		}

		// Clear buffer
		logBuffers[t][label] = make([]string, 0)
	}

	return nil
}

func getLogPathToWrite(t LogType, label string, rowsSize int) (string, bool, error) {
	var (
		filename = ""
		new      = false
	)

	// Get newest file in destination
	logParentDir := fmt.Sprintf("%s/%s", logBaseDir, logTypes[t])
	dir, err := os.ReadDir(logParentDir)
	if err != nil {
		return filename, new, errors.Wrapf(err, "failed to get parent dir for log type: %s", logTypes[t])
	}
	var (
		newestFilenameTime = time.Time{}
		newestFileDirent   os.DirEntry
	)
	for _, dirEnt := range dir {
		if dirEnt.Type().IsRegular() {
			var (
				fname              = dirEnt.Name()
				fnameParts         = strings.Split(fname, "_")
				maybeMatchingLabel = strings.Join(fnameParts[:len(fnameParts)-1], "_")
			)
			if label != maybeMatchingLabel {
				continue
			}

			t, err := time.Parse(FNAME_TIME_FORMAT, fnameParts[len(fnameParts)-1])
			if err != nil {
				continue
			}

			if t.After(newestFilenameTime) {
				newestFilenameTime = t
				newestFileDirent = dirEnt
			}
		}
	}

	new = newestFilenameTime.IsZero()
	if !new {
		fs, err := newestFileDirent.Info()
		if err != nil {
			return filename, new, errors.Errorf("failed to state existing found log file: %s", newestFileDirent.Name())
		}

		if (fs.Size() + int64(rowsSize)) > LOG_FSIZE_LIMIT {
			new = true
		}
	}

	filename = fmt.Sprintf("%s_%s", label, time.Now().Format(FNAME_TIME_FORMAT))
	if !new {
		filename = newestFileDirent.Name()
	}
	path := fmt.Sprintf("%s/%s", logParentDir, filename)
	return path, new, nil
}

func getSizeOfBuffer(buffer []string) int {
	size := 0
	for _, l := range buffer {
		size += len(l) + 1
	}
	return size
}
