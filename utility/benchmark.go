package utility

import (
	"fmt"
	"time"
)

// PrintTook, provide a pointer to either `start` or `took` to print a duration line
func PrintTookStart(prepend string, start *time.Time) {
	if start == nil {
		return
	}
	fmt.Printf("%s, took: %v\n", prepend, time.Since(*start))
}

func PrintTookTook(prepend string, took *time.Duration) {
	if took == nil {
		return
	}
	fmt.Printf("%s, took: %v\n", prepend, *took)
}
