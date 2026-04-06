package utility

import (
	"fmt"
	"time"
)

func PrintTook(prepend string, start *time.Time) {
	if start == nil {
		return
	}
	took := time.Since(*start)
	fmt.Printf("%s, took: %v\n", prepend, took)
}
