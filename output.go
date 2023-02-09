// Copyright (c) 2023, Geert JM Vanderkelen

package gomake

import (
	"fmt"
	"io"
)

func FExitErrorf(w io.Writer, format string, a ...any) {
	FExitError(w, fmt.Sprintf(format, a...))
}

func FExitError(w io.Writer, a ...any) {
	FPrintError(w, a...)
}

func FPrintError(w io.Writer, a ...any) {
	_, _ = fmt.Fprintln(w, "Error:", fmt.Sprint(a...))
}
