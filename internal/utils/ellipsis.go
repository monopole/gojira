package utils

const ellipsis = "…"

func Ellipsis(v string, size int) string {
	if len(v) <= size {
		return v
	}
	return v[:size-1] + ellipsis
}
