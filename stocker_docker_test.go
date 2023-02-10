// Copyright (c) 2023, Geert JM Vanderkelen

package gomake

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golistic/xt"
)

func TestTargetDockerBuild(t *testing.T) {
	if !haveDocker {
		t.Skip("the 'docker' command is not available in PATH")
	}

	wd, err := filepath.Abs("_test_docker")
	xt.OK(t, err)

	t.Run("target is executed", func(t *testing.T) {
		target := TargetDockerBuild
		target.Flags = map[string]any{"image": "example", "tag": "0.9.0"}
		target.WorkDir = wd

		var bufOut strings.Builder
		var bufErr strings.Builder
		m := NewMaker()
		m.StdOut = &bufOut
		m.StdErr = &bufErr
		m.registerTargets(&target)

		exit := m.make(target.Name)
		dataOut := bufOut.String()
		dataErr := bufErr.String()

		fmt.Println("###", dataOut)

		xt.Eq(t, 1, exit)
		xt.Assert(t, strings.Contains(dataOut, target.PreMessages[0]))
		xt.Assert(t, strings.Contains(dataOut, target.PostMessages[0]))
		xt.Assert(t, strings.Contains(dataErr, "RUN exit 1 # we want `docker build` to fail"))
	})

	t.Run("required flags", func(t *testing.T) {
		cases := []struct {
			missing string
			given   []string
		}{
			{
				missing: "-image",
				given:   []string{"-tag", "0.9.0"},
			},
			{
				missing: "-tag",
				given:   []string{"-image", "example"},
			},
		}

		for _, c := range cases {
			t.Run(c.missing+" is required", func(t *testing.T) {
				exp := fmt.Sprintf(`Error: docker-build: flag %s is required`, c.missing)

				var bufErr strings.Builder
				m := NewMaker()
				m.StdErr = &bufErr
				m.registerTargets(&TargetDockerBuild)

				xt.Eq(t, 1, m.make(append([]string{TargetDockerBuild.Name}, c.given...)...))
				xt.Eq(t, exp, strings.TrimSpace(bufErr.String()))
			})
		}
	})
}

func TestTargetDockerBuildX(t *testing.T) {
	if !haveDocker {
		t.Skip("the 'docker' command is not available in PATH")
	}

	wd, err := filepath.Abs("_test_docker")
	xt.OK(t, err)

	t.Run("target is executed", func(t *testing.T) {
		target := TargetDockerBuildXPush
		target.Flags = map[string]any{"image": "example", "registry": "fake.example.com", "tag": "0.9.0"}
		target.WorkDir = wd

		var bufOut strings.Builder
		var bufErr strings.Builder
		m := NewMaker()
		m.StdOut = &bufOut
		m.StdErr = &bufErr
		m.registerTargets(&target)

		exit := m.make(target.Name)
		dataOut := bufOut.String()
		dataErr := bufErr.String()

		xt.Eq(t, 1, exit)

		xt.Assert(t, strings.Contains(dataOut, target.PreMessages[0]))
		xt.Assert(t, strings.Contains(dataErr, "RUN exit 1 # we want `docker build` to fail"))
	})

	t.Run("required flags", func(t *testing.T) {
		cases := []struct {
			missing string
			given   []string
		}{
			{
				missing: "image",
				given:   []string{"-registry", "fake.example.com", "-tag", "0.9.0"},
			},
			{
				missing: "registry",
				given:   []string{"-image", "example", "-tag", "0.9.0"},
			},
			{
				missing: "tag",
				given:   []string{"-image", "example", "-registry", "fake.example.com"},
			},
		}

		for _, c := range cases {
			t.Run(c.missing+" is required", func(t *testing.T) {
				target := TargetDockerBuildXPush
				exp := fmt.Sprintf(`Error: docker-buildx: flag -%s is required`, c.missing)

				var bufErr strings.Builder
				m := NewMaker()
				m.StdErr = &bufErr
				m.registerTargets(&target)

				xt.Eq(t, 1, m.make(append([]string{target.Name}, c.given...)...))
				xt.Eq(t, exp, strings.TrimSpace(bufErr.String()))
			})
		}
	})
}
