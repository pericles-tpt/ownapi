package main

import (
	"fmt"
	"strings"
	"time"
)

const (
	garminDateTimeFormat = "2006-01-02-15-04-05"
	outputDateTimeFormat = "Mon Jan 02 2006 3:04PM"
)

// ActivityListToTimestamps, converts a list of activities from a Garmin
//
// to a summary of uploaded activity times
func ActivityListToSummary(activities []string) (string, error) {
	var (
		invalidActivitesNames = make([]string, 0, len(activities))
		validActivityTimes    = make([]string, 0, len(activities))
	)
	for _, act := range activities {
		t, err := getTime(act)
		if err != nil {
			invalidActivitesNames = append(invalidActivitesNames, act)
			continue
		}
		validActivityTimes = append(validActivityTimes, t.Format(outputDateTimeFormat))
	}
	if len(invalidActivitesNames) > 0 {
		return "", fmt.Errorf("failed to parse activities with the following pre-extension names:\n\t- %s", strings.Join(invalidActivitesNames, "\n\t- "))
	}

	return fmt.Sprintf("Successfuly imported garmin activities started at the following times: %s", strings.Join(validActivityTimes, ", ")), nil
}

func getTime(name string) (time.Time, error) {
	return time.Parse(garminDateTimeFormat, strings.Split(name, ".")[0])
}
