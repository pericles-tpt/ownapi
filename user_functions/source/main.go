package main

import (
	"fmt"

	"github.com/pericles-tpt/ownapi/user_functions/source/garmin"
)

// NOTE: This file will be ignored for the purposes of file generation, feel free to use main() to test out your functions
//
// You can also add utilities here that you don't want included in the generated code
func main() {
	s, err := garmin.ActivityListToSummary([]string{"hello"})
	if err != nil {
		panic(err)
	}
	fmt.Println(s)
}
