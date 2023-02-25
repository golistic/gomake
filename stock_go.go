// Copyright (c) 2023, Geert JM Vanderkelen

package gomake

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func targetVendorHandleFlags(target *Target) (*flag.FlagSet, error) {
	flagSet := flag.NewFlagSet(target.Name, flag.ExitOnError)
	if target.Flags == nil {
		target.Flags = map[string]any{}
	}

	var out string
	flagSet.StringVar(&out, "out", "vendor", "create vendor directory at given path")

	if err := flagSet.Parse(target.FlagArgs); err != nil {
		return nil, err
	}

	if v, ok := target.Flags["out"]; !ok || v == "" {
		target.Flags["out"] = out
	}

	return flagSet, nil
}

var TargetVendor = Target{
	Name:         "vendor",
	PreMessages:  []string{"running go mod vendor command"},
	PostMessages: []string{"done running go mod vendor command"},
	HandleFlags:  targetVendorHandleFlags,
	Do: func(target *Target) error {
		if _, err := target.HandleFlags(target); err != nil {
			return err
		}
		out, ok := target.Flags["out"]

		execArgs := []string{"mod", "vendor"}
		if ok {
			if o, ok := out.(string); ok {
				execArgs = append(execArgs, "-o", o)
			} else {
				return fmt.Errorf("invalid vendor output folder; was %#v", out)
			}
		}
		cmd := exec.Command("go", execArgs...)
		if err := cmd.Run(); err != nil {
			return err
		}
		return nil
	},
}

var TargetCleanupVendor = Target{
	Name:         "clean-vendor",
	PreMessages:  []string{"removing vendor folder"},
	PostMessages: []string{"done removing vendor folder"},
	HandleFlags:  targetVendorHandleFlags,
	Do: func(target *Target) error {
		if _, err := target.HandleFlags(target); err != nil {
			return err
		}
		vendorPath := "vendor"
		out, ok := target.Flags["out"]

		if ok {
			if o, ok := out.(string); ok {
				vendorPath = o
			} else {
				return fmt.Errorf("invalid vendor folder; was %#v", out)
			}
		}

		return os.RemoveAll(vendorPath)
	},
}

// TargetGoVersion is available, but it is just useful for testing.
var TargetGoVersion = Target{
	Name:         "go-version",
	PreMessages:  []string{"running go version"},
	PostMessages: []string{"done running go version"},
	Do: func(target *Target) error {
		var buf strings.Builder

		cmd := exec.Command("go", "version")
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		if err := cmd.Run(); err != nil {
			return err
		}

		target.Maker.Println(strings.TrimSpace(buf.String()))
		return nil
	},
}

// TargetGoLint runs golangci-lint executing lots of linters against the project's source code
var TargetGoLint = Target{
	Name:         "go-lint",
	Description:  "Runs golangci-lint to executing various linters against the projects Go source.",
	PreMessages:  []string{"running golangci-lint"},
	PostMessages: []string{"done running golangci-lint"},
	Do: func(target *Target) error {
		var bufOut strings.Builder
		var bufErr strings.Builder

		cmd := exec.Command("golangci-lint", "run", "--color", "always", "./...")
		cmd.Stdout = &bufOut
		cmd.Stderr = &bufErr

		if err := cmd.Start(); err != nil {
			return err
		}

		err := cmd.Wait()
		if err != nil {
			switch err.(type) {
			case *exec.ExitError:
				fmt.Println(bufOut.String())
				fmt.Println(bufErr.String())
				return nil
			default:
				return err
			}
		}

		target.Maker.Println("Congrats! Looking good!")
		return nil
	},
}
