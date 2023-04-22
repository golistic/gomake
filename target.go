// Copyright (c) 2023, Geert JM Vanderkelen

package gomake

import (
	"flag"
)

type Target struct {
	Maker *Maker

	Name            string
	Description     string
	FlagArgs        []string
	Flags           map[string]any
	HandleFlags     func(target *Target) (*flag.FlagSet, error)
	PreMessages     []string
	PostMessages    []string
	DeferredTargets []*Target
	PreTargets      []*Target
	Do              func(*Target) error
	WorkDir         string
	Settings        map[string]any
}
