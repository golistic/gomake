// Copyright (c) 2023, Geert JM Vanderkelen

package gomake

import (
	"fmt"
	"io"
	"os"
)

type Maker struct {
	StdOut         io.Writer
	StdErr         io.Writer
	targetRegistry map[string]*Target
	msgPrefix      string
}

func NewMaker() *Maker {
	return &Maker{
		StdOut:         os.Stdout,
		StdErr:         os.Stderr,
		targetRegistry: map[string]*Target{},
		msgPrefix:      "==>",
	}
}

func (m *Maker) make(args ...string) int {
	if len(m.targetRegistry) == 0 {
		m.PrintlnError("no targets available")
		return 1
	}

	if len(args) == 0 {
		m.Print(helpAvailableTargets(m))
		return 0
	}

	targetName := args[0]

	if targetName == "help" {
		m.Print(helpAvailableTargets(m))
		return 0
	}

	target, ok := m.targetRegistry[targetName]
	if !ok {
		m.PrintfError("target %s not available\n\n%s\n", targetName, helpAvailableTargets(m))
		return 1
	}

	if len(args) >= 1 {
		target.FlagArgs = args[1:]
	}

	return m.run(target)
}

func (m *Maker) registerTargets(targets ...*Target) {
	for _, target := range targets {
		_, ok := m.targetRegistry[target.Name]
		if ok {
			FExitErrorf(m.StdErr, "target %s cannot be registered more than once; was ", target.Name)
		}
		m.targetRegistry[target.Name] = target
	}
}

func (m *Maker) run(targets ...*Target) int {
	for _, target := range targets {
		if ret := m.runTarget(target); ret > 0 {
			return ret
		}
	}

	return 0
}

func (m *Maker) runTarget(target *Target) int {
	target.Maker = m
	if target.HandleFlags != nil {
		if _, err := target.HandleFlags(target); err != nil {
			m.PrintError(err)
			return 0
		}
	}

	for _, msg := range target.PreMessages {
		m.Println(m.msgPrefix, msg)
	}

	defer func() {
		m.run(target.DeferredTargets...)
	}()

	if ret := m.run(target.PreTargets...); ret > 0 {
		return ret
	}

	if err := target.Do(target); err != nil {
		m.PrintError(err)
		return 1
	}
	for _, msg := range target.PostMessages {
		m.Println(m.msgPrefix, msg)
	}

	return 0
}

func (m *Maker) PrintfError(format string, a ...any) {
	_, _ = fmt.Fprintf(m.StdErr, "Error: "+format, a...)
}

func (m *Maker) PrintError(a ...any) {
	_, _ = fmt.Fprint(m.StdErr, append([]any{"Error:"}, a...)...)
}

func (m *Maker) PrintlnError(a ...any) {
	_, _ = fmt.Fprintln(m.StdErr, append([]any{"Error:"}, a...)...)
}

func (m *Maker) Printf(format string, a ...any) {
	_, _ = fmt.Fprintf(m.StdOut, format, a...)
}

func (m *Maker) Print(a ...any) {
	_, _ = fmt.Fprint(m.StdOut, a...)
}

func (m *Maker) Println(a ...any) {
	_, _ = fmt.Fprintln(m.StdOut, a...)
}
