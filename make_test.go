// Copyright (c) 2023, Geert JM Vanderkelen

package gomake

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/golistic/xt"
)

func excludePrefixedMessages(prefix, lines string) []string {
	var res []string
	for _, s := range strings.Split(lines, "\n") {
		if !strings.HasPrefix(s, prefix) {
			res = append(res, s)
		}
	}
	return res
}

func TestMake(t *testing.T) {
	t.Run("no targets registered", func(t *testing.T) {
		exp := `Error: no targets available
`
		var buf strings.Builder
		m := NewMaker()
		m.StdErr = &buf

		xt.Eq(t, 1, m.make())
		xt.Eq(t, exp, buf.String())
	})

	t.Run("targets have been registered", func(t *testing.T) {
		exp := `Available targets:
   go-version
   vendor
`
		var buf strings.Builder
		m := NewMaker()
		m.StdOut = &buf

		m.registerTargets(&TargetGoVersion, &TargetVendor)
		xt.Eq(t, 0, m.make())
		xt.Eq(t, exp, buf.String())
	})

	t.Run("run a target", func(t *testing.T) {
		exp := fmt.Sprintf("go version %s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

		var buf strings.Builder
		m := NewMaker()
		m.StdOut = &buf
		m.registerTargets(&TargetGoVersion, &TargetVendor)

		xt.Eq(t, 0, m.make("go-version"))
		have := excludePrefixedMessages(m.msgPrefix, buf.String())
		xt.Assert(t, len(have) != 0, "expected output from target")
		xt.Eq(t, exp, have[0])
	})
}
