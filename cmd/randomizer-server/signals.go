//go:build !unix

package main

import "os"

var signals = []os.Signal{os.Interrupt}
