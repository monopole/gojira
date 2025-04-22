package myj

//go:generate go run github.com/dmarkham/enumer -linecomment -type=Rel
type Rel int

const (
	RelUnknown     Rel = iota
	RelEqual           // =
	RelNotEqual        // !=
	RelLess            // <
	RelLessOrEqual     // <=
)
