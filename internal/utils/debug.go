package utils

import (
	"github.com/spf13/pflag"
)

// Debug is a global var to toggle debugging.  Amateurish.
var Debug bool

func FlagsAddDebug(set *pflag.FlagSet) {
	set.BoolVar(&Debug, "debug", false, "enable printing of debugging info")
}
