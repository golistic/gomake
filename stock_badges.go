// Copyright (c) 2023, Geert JM Vanderkelen

package gomake

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/golistic/shieldbadger"
)

// TargetBadges creates badges using Shields.io suitable for showing on, for example,
// GitHub. Badges are stored by default the folder `.badges` in the root of the
// repository.
// The definition of the badges is stored in JSON file `_badges/badges.json`.
var TargetBadges = Target{
	Name:         "badges",
	PreMessages:  []string{"generating badges"},
	PostMessages: []string{"done generating badges"},
	HandleFlags: func(target *Target) (*flag.FlagSet, error) {
		flagSet := flag.NewFlagSet(target.Name, flag.ExitOnError)
		if target.Flags == nil {
			target.Flags = map[string]any{}
		}

		var (
			configFile string
			destFolder string
		)

		flagSet.StringVar(&configFile, "config", filepath.Join("_badges", "badge.json"),
			"Configuration file containing which badges to generate")
		flagSet.StringVar(&destFolder, "dest", "_badges/",
			"Folder in which fetched badges will be stored")

		if err := flagSet.Parse(target.FlagArgs); err != nil {
			return nil, err
		}

		if configFile != "" {
			target.Flags["config"] = configFile
		} else {
			return nil, fmt.Errorf("config file is required")
		}

		if configFile != "" {
			target.Flags["destFolder"] = destFolder
		} else {
			return nil, fmt.Errorf("destination folder is required")
		}

		return flagSet, nil
	},
	Do: func(target *Target) error {
		if _, err := target.HandleFlags(target); err != nil {
			return err
		}

		sb, err := shieldbadger.NewShieldBadger(target.Flags["config"].(string))
		if err != nil {
			return err
		}

		sb.FeedbackCallback = func(format string, a ...any) {
			target.Maker.Printf(format, a...)
		}

		return sb.Fetch()
	},
}
