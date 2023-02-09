// Copyright (c) 2023, Geert JM Vanderkelen

package gomake

import (
	"flag"
	"os"
)

var defaultMake = NewMaker()

func Make() {
	flag.Parse()
	os.Exit(defaultMake.make(flag.Args()...))
}

func RegisterTargets(targets ...*Target) {
	defaultMake.registerTargets(targets...)
}
