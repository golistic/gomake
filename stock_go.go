// Copyright (c) 2023, Geert JM Vanderkelen

package gomake

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
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

// TargetGoCoverage runs Go tests and retrieves the coverage report.
var TargetGoCoverage = Target{
	Name:         "go-coverage",
	Description:  "Runs both unittests and integration tests to calculate coverage.",
	PreMessages:  []string{"running go coverage"},
	PostMessages: []string{"done running coverage"},
	Settings: map[string]any{
		"integration": [][]string{
			// cannot at go-coverage (would recursively run)
			{"go", "run", "-cover", "./cmd/make", "go-version"},
			{"go", "run", "-cover", "./cmd/make", "badges"},
			{"go", "run", "-cover", "./cmd/make", "go-lint"},
		},
	},
	HandleFlags: func(target *Target) (*flag.FlagSet, error) {
		flagSet := flag.NewFlagSet(target.Name, flag.ExitOnError)
		if target.Flags == nil {
			target.Flags = map[string]any{}
		}

		var coverdir string

		flagSet.StringVar(&coverdir, "coverdir", "",
			"Where to store coverage profiles (default: create a system temporary directory)")

		if err := flagSet.Parse(target.FlagArgs); err != nil {
			return nil, err
		}

		if v, ok := target.Flags["coverdir"]; !ok || v == "" {
			target.Flags["coverdir"] = coverdir
		}

		return flagSet, nil
	},
	Do: func(target *Target) error {
		coverDir, _ := target.Flags["coverdir"].(string) // when missing/incorrect, we use temporary
		integration, ok := target.Settings["integration"].([][]string)
		if !ok {
			return fmt.Errorf("integration setting not slice of string slices")
		}

		result, err := combinedCoverage(target.Maker, coverDir, integration)
		if err != nil {
			return err
		}

		target.Maker.Println("Total Coverage:", result)
		return nil
	},
}

func combinedCoverage(maker *Maker, coverDir string, integration [][]string) (string, error) {
	var bufErr strings.Builder

	if strings.TrimSpace(coverDir) == "" {
		var err error
		coverDir, err = os.MkdirTemp("", "gomake-go-coverage")
		if err != nil {
			return "", err
		}
		defer func() { _ = os.RemoveAll(coverDir) }()
	} else {
		if err := os.Mkdir(coverDir, 0770); err != nil {
			switch {
			case os.IsExist(err):
				fmt.Println("Coverage output directory exists; you are responsible to clean it up before and after")
			case err != nil:
				return "", err
			}
		}
	}

	fmt.Println("Coverage profiles stored in", coverDir)

	dirUnit := path.Join(coverDir, "unittests")
	if err := os.Mkdir(dirUnit, 0700); err != nil {
		return "", err
	}
	dirIntegration := path.Join(coverDir, "integration")
	if err := os.Mkdir(dirIntegration, 0700); err != nil {
		return "", err
	}

	env := os.Environ()
	env = append(env, "GOCOVERDIR="+dirIntegration)

	maker.Println("Coverage using unittests")
	cmd := []string{"go", "test", "-cover", "./...",
		"-args", fmt.Sprintf("-test.gocoverdir=%s", dirUnit)}
	if err := execCmd(nil, &bufErr, nil, cmd...); err != nil {
		switch err.(type) {
		case *exec.ExitError:
			fmt.Println(bufErr.String())
			return "", nil
		default:
			return "", err
		}
	}

	if len(integration) > 0 {
		maker.Println("Coverage using integration")
	}

	for _, cmdAndArgs := range integration {
		fmt.Println("  Running:", strings.Join(cmdAndArgs, " "))
		var bufErr strings.Builder
		if err := execCmd(nil, &bufErr, env, cmdAndArgs...); err != nil {
			switch err.(type) {
			case *exec.ExitError:
				fmt.Println(bufErr.String())
				return "", nil
			default:
				return "", err
			}
		}
	}

	bufOut := strings.Builder{}
	cmdAndArgs := []string{
		"go", "tool", "covdata", "textfmt",
		"-i", dirIntegration + "," + dirUnit,
		"-o", path.Join(coverDir, "profile"),
	}
	if err := execCmd(&bufOut, &bufErr, nil, cmdAndArgs...); err != nil {
		switch err.(type) {
		case *exec.ExitError:
			fmt.Println(bufErr.String())
			return "", nil
		default:
			return "", err
		}
	}

	bufOut = strings.Builder{}
	cmdAndArgs = []string{
		"go", "tool", "cover", "-func", path.Join(coverDir, "profile"),
	}
	if err := execCmd(&bufOut, &bufErr, nil, cmdAndArgs...); err != nil {
		switch err.(type) {
		case *exec.ExitError:
			fmt.Println(bufErr.String())
			return "", nil
		default:
			return "", err
		}
	}

	reTotal := regexp.MustCompile(`total:\s+\(\w+\)\s+(.+?)%\n`)
	m := reTotal.FindStringSubmatch(bufOut.String())
	if m == nil {
		return "", fmt.Errorf("parsing output of Go cover tool (getting total)")
	}

	return m[1], nil
}
