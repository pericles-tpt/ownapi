package garmin

import (
	"fmt"
	"strings"
	"time"
)

const (
	garminDateTimeFormat = "2006-01-02-15-04-05"
	outputDateTimeFormat = "Mon Jan 02 2006 3:04PM"
)

type Thing1 struct {
	A int `json:"a"`
	B string
	C bool
}

// ActivityListToTimestamps, converts a list of activities from a Garmin
//
// to a summary of uploaded activity times yo
func ActivityListToSummary(activities []string) (string, error) {
	var (
		invalidActivitesNames = make([]string, 0, len(activities))
		validActivityTimes    = make([]string, 0, len(activities))
	)

	if len(activities) == 0 {
		return "", nil
	}

	for _, act := range activities {
		pathParts := strings.Split(act, "/")
		basename := pathParts[len(pathParts)-1]

		basenameParts := strings.Split(basename, ".")
		if len(basenameParts) > 1 && basenameParts[1] == "fit" {
			datetimePart := basenameParts[0]
			t, err := getTime(datetimePart)
			if err != nil {
				invalidActivitesNames = append(invalidActivitesNames, act)
				continue
			}
			validActivityTimes = append(validActivityTimes, t.Format(outputDateTimeFormat))
		}
	}

	return fmt.Sprintf("Successfuly imported garmin activities started at the following times: %s", strings.Join(validActivityTimes, ", ")), nil
}

// a single line comment
/*
	a multiline comment
	followed
	by
*/
func getTime(datetimePart string) (time.Time, error) {
	return time.Parse(garminDateTimeFormat, datetimePart)
}

type Thing2 struct {
	A int `json:"a"`
	B string
	C bool
}

/*
	a multiline comment
	followed
	by
*/
// a single line comment
func bad() int {
	return 32
}

// something else
/*
	a multiline comment
	followed
	by
*/
// a single line comment
func haha() bool {
	return false
}

/*
	a multiline comment
	followed
	by
*/
// a single line comment
/*
	a multiline comment
	followed
	by
*/
func foo() {

}
