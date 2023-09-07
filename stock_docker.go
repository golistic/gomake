// Copyright (c) 2023, Geert JM Vanderkelen

package gomake

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"os/exec"
	"strings"
)

var TargetDockerBuild = Target{
	Name:            "docker-build",
	FlagArgs:        nil,
	Flags:           nil,
	PreMessages:     []string{"building image"},
	PostMessages:    []string{"done building Docker image"},
	DeferredTargets: nil,
	PreTargets:      nil,
	HandleFlags: func(target *Target) (*flag.FlagSet, error) {
		flagSet := flag.NewFlagSet(target.Name, flag.ExitOnError)
		if target.Flags == nil {
			target.Flags = map[string]any{}
		}

		var (
			registry string
			image    string
			tag      string
		)

		flagSet.StringVar(&registry, "registry", "",
			"Docker registry to be used when naming the image")
		flagSet.StringVar(&image, "image", "", "Docker image name")
		flagSet.StringVar(&tag, "tag", "", "Docker Image tag (usually version)")

		if err := flagSet.Parse(target.FlagArgs); err != nil {
			return nil, err
		}

		if registry != "" {
			// command line flag override default
			if registry == "docker.io" || registry == "local" {
				registry = "docker.io"
			}

			target.Flags["registry"] = registry
		}

		if v, ok := target.Flags["registry"]; !ok || v == "" {
			fmt.Println("Note: registry not set, default docker.io/library will be used.")
		}

		if target.Flags["registry"] == "docker.io" {
			// correct image name (checked later when empty)
			if image, ok := target.Flags["image"].(string); ok && image != "" {
				target.Flags["image"] = "library/" + image[strings.LastIndex(image, "/")+1:]
			}
		} else if image != "" {
			// command line flag override default
			target.Flags["image"] = image
		}

		if v, ok := target.Flags["image"]; !ok || v == "" {
			if image == "" {
				return nil, fmt.Errorf("%s: flag -image is required", target.Name)
			}
		}

		if tag != "" {
			// command line flag override default
			target.Flags["tag"] = tag
		}

		if v, ok := target.Flags["tag"]; !ok || v == "" {
			if tag == "" {
				return nil, fmt.Errorf("%s: flag -tag is required", target.Name)
			}
		}

		return flagSet, nil
	},
	Do: func(target *Target) error {
		if _, err := target.HandleFlags(target); err != nil {
			return err
		}

		tag := fmt.Sprintf("%s:%s", target.Flags["image"].(string), target.Flags["tag"].(string))

		if r, ok := target.Flags["registry"]; ok {
			var err error
			tag, err = url.JoinPath(r.(string), tag)
			if err != nil {
				return fmt.Errorf("failed creating tag using registry (%w)", err)
			}
		}

		execArgs := []string{"build", "--tag", tag, "."}
		if err := execDocker(target.Maker.StdOut, target.Maker.StdErr, execArgs, target.WorkDir); err != nil {
			return err
		}

		return nil
	},
}

var TargetDockerBuildXPush = Target{
	Name:            "docker-buildx",
	FlagArgs:        nil,
	Flags:           nil,
	PreMessages:     []string{"building image"},
	PostMessages:    []string{"done building Docker image"},
	DeferredTargets: nil,
	PreTargets:      nil,
	HandleFlags: func(target *Target) (*flag.FlagSet, error) {
		flagSet := flag.NewFlagSet(target.Name, flag.ExitOnError)
		if target.Flags == nil {
			target.Flags = map[string]any{}
		}

		var (
			registry string
			image    string
			tag      string
			platform string
		)

		flagSet.StringVar(&registry, "registry", "",
			"Docker registry to push too (authentication must be done before)")
		flagSet.StringVar(&image, "image", "", "Docker image name")
		flagSet.StringVar(&tag, "tag", "", "Docker image tag (usually version)")
		flagSet.StringVar(&platform, "platform", "linux/arm64,linux/amd64",
			"Platforms to build for (comma separated)")

		if err := flagSet.Parse(target.FlagArgs); err != nil {
			return nil, err
		}

		if v, ok := target.Flags["registry"]; !ok || v == "" {
			if registry == "" {
				return nil, fmt.Errorf("%s: flag -registry is required", target.Name)
			}
			target.Flags["registry"] = registry
		}

		if v, ok := target.Flags["image"]; !ok || v == "" {
			if image == "" {
				return nil, fmt.Errorf("%s: flag -image is required", target.Name)
			}
			target.Flags["image"] = image
		}

		if v, ok := target.Flags["tag"]; !ok || v == "" {
			if tag == "" {
				return nil, fmt.Errorf("%s: flag -tag is required", target.Name)
			}
			target.Flags["tag"] = tag
		}

		if v, ok := target.Flags["platform"]; !ok || v == "" {
			target.Flags["platform"] = platform
		}

		return flagSet, nil
	},
	Do: func(target *Target) error {
		tag := fmt.Sprintf("%s:%s", target.Flags["image"].(string), target.Flags["tag"].(string))
		fullTag, err := url.JoinPath(target.Flags["registry"].(string), tag)
		if err != nil {
			return fmt.Errorf("failed creating tag using registry (%w)", err)
		}

		builderName := fmt.Sprintf("gomake-tempoary-%d", rand.Int())
		// create a builder
		execArgs := []string{
			"buildx", "create",
			"--name", builderName,
			"--driver", "docker-container",
			"--use",
		}
		if err := execDocker(target.Maker.StdOut, target.Maker.StdErr, execArgs, target.WorkDir); err != nil {
			return err
		}

		defer func() {
			execArgs := []string{
				"buildx", "rm", "-f", builderName,
			}
			_ = execDocker(target.Maker.StdOut, target.Maker.StdErr, execArgs, target.WorkDir)
		}()

		execArgs = []string{
			"buildx", "build",
			"--builder", builderName,
			"--platform", target.Flags["platform"].(string),
			"--tag", fullTag, "--push", ".",
		}
		if err := execDocker(target.Maker.StdOut, target.Maker.StdErr, execArgs, target.WorkDir); err != nil {
			return err
		}

		return nil
	},
}

func execDocker(stdOut, stdErr io.Writer, args []string, workDir string) error {
	cmd := exec.Command("docker", args...)
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr
	if workDir != "" {
		fmt.Println("executing in directory:", workDir)
		cmd.Dir = workDir
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
