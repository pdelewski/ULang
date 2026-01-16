package compiler

import (
	"fmt"
	"log"
)

// DebugMode controls whether debug output is enabled
var DebugMode bool

// DebugPrintf prints only when debug mode is enabled
func DebugPrintf(format string, args ...interface{}) {
	if DebugMode {
		fmt.Printf(format, args...)
	}
}

// DebugLogPrintf logs only when debug mode is enabled
func DebugLogPrintf(format string, args ...interface{}) {
	if DebugMode {
		log.Printf(format, args...)
	}
}
