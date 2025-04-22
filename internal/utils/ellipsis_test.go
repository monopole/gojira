package utils

import (
	"testing"
)

func TestEllipsis(t *testing.T) {
	tests := map[string]struct {
		v    string
		size int
		want string
	}{
		"t2": {
			v:    "aaabbb",
			size: 6,
			want: "aaabbb",
		},
		"t3": {
			v:    "aaabbb",
			size: 5,
			want: "aaab…",
		},
		"t4": {
			v:    "aaabbb",
			size: 3,
			want: "aa…",
		},
	}
	for n, tc := range tests {
		t.Run(n, func(t *testing.T) {
			if got := Ellipsis(tc.v, tc.size); got != tc.want {
				t.Errorf("Ellipsis() = %v, want %v", got, tc.want)
			}
		})
	}
}
