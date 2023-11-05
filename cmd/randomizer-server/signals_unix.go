//go:build unix

package main

import (
	"os"
	"syscall"
)

var exitSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}
