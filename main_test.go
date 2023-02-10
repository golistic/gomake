// Copyright (c) 2023, Geert JM Vanderkelen

package gomake

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

var (
	testExitCode int
	testErr      error
)

var (
	haveDocker bool
)

func testTearDown() {
	if testErr != nil {
		testExitCode = 1
		fmt.Println(testErr)
	}
}

func TestMain(m *testing.M) {
	defer func() { os.Exit(testExitCode) }()
	defer testTearDown()

	if _, err := exec.LookPath("docker"); err == nil {
		haveDocker = true
	}

	testExitCode = m.Run()
}
