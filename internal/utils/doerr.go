package utils

import (
	"fmt"
	"os"
)

func DoErrF(f string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, f, args...)
}

func DoErr1(f ...string) {
	_, _ = fmt.Fprintln(os.Stderr, f)
}
