package utility

import (
	"syscall"
	"time"
)

const (
	// Magic number to improve accuracy for sleep time
	MIN_SLEEP_ACCURACY int64 = 50869
	BILLION            int64 = 1_000_000_000
)

// SleepLinux, uses the nanosleep syscall to improve accuracy of sleep over the default time.Sleep() as of go 1.25
// it's pretty accurate down to ~50us, below that not so much
func SleepLinux(sleepFor time.Duration) {
	// HACK: For microsecond accuracy time.Sleep() is woefully slow on linux, syscall.Nanosleep is better but still
	// 		 requires an `AUTO_RUN_LOOP_WAIT_OFFSET` I derived from a trial and error process.
	//
	// 		 With this setup a 250us loop interval is += 2.5% of the target finish time ~99% of the time which I
	// 		 think is acceptable. Accuracy improves for longer intervals and likely gets worse for shorter intervals.
	//
	//		 Weirdly the absolute difference between the "expected" finish and "actual" finish does increase for much
	//		 longer intervals like 75ms, but as a percentage of the overall autoRunLoopFrequency it's much improved.
	// 		 This might be related to some sort of caching that occurs for `Nanosleep`s that occur within microseconds
	//		 of one another on linux?
	//
	//		 In practice the interval mightn't be in the microseconds anyway, since it's dynamically retrieved from
	// 	     `getAutoRunLoopFrequency` and is relative to the smallest interval defined on a pipeline.
	var (
		totalSleep       = time.Duration(sleepFor.Nanoseconds() - MIN_SLEEP_ACCURACY).Nanoseconds()
		totalSleepNSPart = totalSleep % BILLION
		totalSleepSPart  = ((totalSleep - totalSleepNSPart) / BILLION)
		ts               = &syscall.Timespec{Nsec: totalSleepNSPart, Sec: totalSleepSPart}
		err              = syscall.Nanosleep(ts, nil)
	)
	if err != nil {
		// Can return "interrupted system call" error, retry once always on error
		syscall.Nanosleep(ts, nil)
	}
}

func SleepLinuxUntilIteration(start time.Time, iteration int64, interval time.Duration) {
	sleepFor := time.Until(start.Add(interval * time.Duration(iteration)))
	SleepLinux(sleepFor)
}
