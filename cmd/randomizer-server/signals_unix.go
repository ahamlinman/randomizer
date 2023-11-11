//go:build unix

package main

import "syscall"

func init() {
	exitSignals = append(exitSignals, syscall.SIGTERM)
}
