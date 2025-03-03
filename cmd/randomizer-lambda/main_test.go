package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBannedPackages(t *testing.T) {
	// According to https://github.com/spf13/cobra/blob/v1.9.0/cobra_test.go#L247,
	// `go tool nm` doesn't work on Windows.
	if runtime.GOOS == "windows" {
		t.SkipNow()
	}

	const name = "randomizer-lambda"
	binpath := filepath.Join(t.TempDir(), name)
	build := exec.Command("go", "build", "-tags=lambda.norpc,grpcnotrace", "-o", binpath, ".")
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	if err := build.Run(); err != nil {
		t.Fatalf("failed to build: %v", err)
	}

	var nmout strings.Builder
	nm := exec.Command("go", "tool", "nm", binpath)
	nm.Stdout, nm.Stderr = &nmout, os.Stderr
	if err := nm.Run(); err != nil {
		t.Fatalf("failed to run go tool nm: %v", err)
	}

	if strings.Contains(nmout.String(), "T text/template.") {
		t.Errorf("%s imports text/template, blocking dead code elimination", name)
	}
	if strings.Contains(nmout.String(), "T html/template.") {
		t.Errorf("%s imports html/template, blocking dead code elimination", name)
	}
	if strings.Contains(nmout.String(), "T cloud.google.com/") {
		t.Errorf("%s contains Google Cloud packages, even though it's for AWS", name)
	}
	if strings.Contains(nmout.String(), "T go.etcd.io/bbolt.") {
		t.Errorf("%s imports go.etcd.io/bbolt, even though it's for AWS", name)
	}
}
