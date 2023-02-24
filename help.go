// Copyright (c) 2023, Geert JM Vanderkelen

package gomake

import "sort"

func helpAvailableTargets(m *Maker) string {
	help := "Available targets:\n"

	var names []string
	for _, target := range m.targetRegistry {
		names = append(names, target.Name)
	}

	sort.Strings(names)
	for _, name := range names {
		target := m.targetRegistry[name]
		help += "   " + name + "\n"
		if target.Description != "" {
			help += "      " + target.Description + "\n"
		}
	}
	return help
}
