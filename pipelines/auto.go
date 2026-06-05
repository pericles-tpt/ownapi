package pipelines

import (
	"math"
	"sort"
	"time"

	log2 "github.com/pericles-tpt/ownapi/log"
	"github.com/pericles-tpt/ownapi/node"
	"github.com/pericles-tpt/ownapi/utility"
)

// NOTE: Probably want this to be 100us MINIMUM, the loop takes < 50us at worst so 100us might
//
//	(barely) be a high enough buffer
const (
	MIN_AUTO_RUN_LOOP_FREQUENCY          = 250 * time.Microsecond
	MAX_AUTO_RUN_LOOP_FREQUENCY_MULTIPLE = 366 * 24 * 3600 * 1000 * 4
)

var (
	COUNTER_LIMIT                int64 = math.MaxInt64 / MIN_AUTO_RUN_LOOP_FREQUENCY.Nanoseconds()
	CORRECT_COUNTER_TIMING_EVERY int64 = 1000

	MIN_AUTO_RUN_LOOP_FREQUENCY_NS   = MIN_AUTO_RUN_LOOP_FREQUENCY.Nanoseconds()
	MIN_AUTO_RUN_LOOP_FREQUENCY_UNIT = MIN_AUTO_RUN_LOOP_FREQUENCY.String()[len(MIN_AUTO_RUN_LOOP_FREQUENCY.String())-1]

	MIN_LOG_AUTO_RUN_LOOP_DURATION_PC    = 0.05
	MIN_LOG_AUTO_RUN_LOOP_DURATION_DENOM = int64(MIN_LOG_AUTO_RUN_LOOP_DURATION_PC * 100.0)
)

func ScheduleAutoTriggeredPipelines(pls []Pipeline) {
	var (
		autoTriggeredPipelines = make([]int, 0, len(pls))
		triggerNodes           = make([]node.Trigger, 0, len(pls))
	)
	for i, pl := range pls {
		triggerNode := pl.Nodes[0][0].GetTrigger()
		if triggerNode != nil {
			autoTriggeredPipelines = append(autoTriggeredPipelines, i)
			triggerNodes = append(triggerNodes, *triggerNode)
		}
	}
	if len(autoTriggeredPipelines) == 0 {
		return
	}

	maxRequiredPrime := int(math.Sqrt(float64(MAX_AUTO_RUN_LOOP_FREQUENCY_MULTIPLE)))
	utility.InitPrimes(maxRequiredPrime)

	// Sort nodes by their trigger intervals
	origNode := make([]node.Trigger, len(triggerNodes))
	copy(origNode, triggerNodes)
	sort.SliceStable(triggerNodes, func(i, j int) bool {
		return origNode[i].EveryN < origNode[j].EveryN
	})
	sort.SliceStable(autoTriggeredPipelines, func(i, j int) bool {
		return origNode[i].EveryN < origNode[j].EveryN
	})

	// Make some buckets for common trigger intervals
	var (
		intervals               = []int{triggerNodes[0].EveryN}
		intervalTriggerNodeIdxs = [][]int{{0}}
	)
	for i := 1; i < len(triggerNodes); i++ {
		ivl := triggerNodes[i].EveryN

		lastIdx := len(intervals) - 1
		if intervals[lastIdx] < ivl {
			intervals = append(intervals, ivl)
			intervalTriggerNodeIdxs = append(intervalTriggerNodeIdxs, []int{i})
		} else {
			intervalTriggerNodeIdxs[lastIdx] = append(intervalTriggerNodeIdxs[lastIdx], i)
		}
	}

	// TODO: How to handle updates to node configs?
	// `autoRunLoopFrequency` is 1/100 of the smallest interval OR MIN_AUTO_RUN_LOOP_FREQUENCY
	autoRunLoopFrequency := getAutoRunLoopFrequency(intervals)
	autoRunLoopFrequencyNS := autoRunLoopFrequency.Nanoseconds()
	autoRunLoopFrequencyNSF64 := float64(autoRunLoopFrequencyNS)

	minPrintDuration := MIN_AUTO_RUN_LOOP_FREQUENCY / time.Duration(MIN_LOG_AUTO_RUN_LOOP_DURATION_DENOM)

	var (
		tookLongerBufferLen          = 1000
		tookLongerThanMinPrintValues = make([][]any, 0, tookLongerBufferLen)
		tookLonger                   = 0

		counter       int64 = 0
		counterResets       = 0

		// TODO: This shouldn't be global, probably should be defined per pipeline
		runPipelinesOnStartup = false
	)
	for len(intervals) > 0 {
		befAutoRunLoop := time.Now()
		for counter = 0; counter < COUNTER_LIMIT; counter++ {
			startLoop := time.Now()
			if counter > 0 || runPipelinesOnStartup {
				for i, ivl := range intervals {
					triggerIntervalNS := int64(ivl) * MIN_AUTO_RUN_LOOP_FREQUENCY_NS
					triggerDue := ((counter * autoRunLoopFrequencyNS) % triggerIntervalNS) < autoRunLoopFrequencyNS
					if triggerDue {
						log2.WriteLogs(log2.Auto, "RUNNING", []string{"INTERVAL", "UNIT"}, [][]any{{intervals[i], MIN_AUTO_RUN_LOOP_FREQUENCY_UNIT}}, true)

						// Run pipelines
						for _, triggerNodeIdx := range intervalTriggerNodeIdxs[i] {
							plIdx := autoTriggeredPipelines[triggerNodeIdx]
							pl := pls[plIdx]

							triggerInterval := time.Duration(triggerIntervalNS)
							go Run(&pl.Name, nil, &triggerInterval)
						}
					}
				}
			}
			logicTook := time.Since(startLoop)

			if logicTook >= minPrintDuration {
				timeToFinish := autoRunLoopFrequency - logicTook

				tookLongerThanMinPrintValues = append(tookLongerThanMinPrintValues, []any{timeToFinish, logicTook, ((autoRunLoopFrequencyNSF64 - float64(timeToFinish.Nanoseconds())) / autoRunLoopFrequencyNSF64) * 100.0})
				tookLonger++

				if tookLonger == tookLongerBufferLen-1 {
					err := log2.WriteLogs(log2.Auto, "SLEPT_FOR", []string{"SLEPT_FOR", "TOOK", "TOOK_PC_BUDGET"}, tookLongerThanMinPrintValues, true)
					if err != nil {
						// TODO:
					}

					tookLongerThanMinPrintValues = tookLongerThanMinPrintValues[:0]
					tookLonger = 0
				}
			}

			utility.SleepLinuxUntilIteration(befAutoRunLoop, counter+1, autoRunLoopFrequency)
		}
		counterResets++
	}
}

// getAutoRunLoopFrequency, gets the GCD of all intervals dividing the result by up to 10 (if possible) to improve accuracy
func getAutoRunLoopFrequency(intervals []int) time.Duration {
	freqNS := MIN_AUTO_RUN_LOOP_FREQUENCY_NS
	freqMultiples, _, err := utility.GCD(intervals, 10)
	if err == nil && freqMultiples > 1 {
		freqNS *= int64(freqMultiples)
	}
	return time.Duration(freqNS)
}
