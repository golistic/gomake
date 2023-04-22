// Copyright (c) 2023, Geert JM Vanderkelen

package main

import "github.com/golistic/gomake"

func main() {
	gomake.RegisterTargets(
		&gomake.TargetBadges,
		&gomake.TargetGoCoverage,
		&gomake.TargetGoLint,
		&gomake.TargetGoVersion,
	)
	gomake.Make()
}
