// Copyright (c) 2023, Geert JM Vanderkelen

package gomake

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
)

var defaultMake = NewMaker()

func Make() {
	flag.Parse()
	os.Exit(defaultMake.make(flag.Args()...))
}

func RegisterTargets(targets ...*Target) {
	defaultMake.registerTargets(targets...)
}

func execCmd(stdOut io.Writer, stdErr io.Writer, env []string, cmdAndArgs ...string) error {
	if len(cmdAndArgs) == 0 {
		return fmt.Errorf("no command provided")
	}
	var args []string
	var name string

	name = cmdAndArgs[0]
	if len(cmdAndArgs) > 1 {
		args = cmdAndArgs[1:]
	}

	cmd := exec.Command(name, args...)
	if stdOut != nil {
		cmd.Stdout = stdOut
	}
	if stdErr != nil {
		cmd.Stderr = stdErr
	}

	cmd.Env = env

	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}
