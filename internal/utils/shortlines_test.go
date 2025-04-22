package utils

import (
	"testing"
)

func Test_ShortLines(t *testing.T) {
	tests := map[string]struct {
		arg  string
		want string
	}{
		"t0": {
			arg: ``,
			want: `
`,
		},
		"t1": {
			arg: `one`,
			want: `
one`,
		},
		"t2": {
			arg: `one two three`,
			want: `
one two three`,
		},
		"t3": {
			arg: `one two three four`,
			want: `
one two three four`,
		},
		"t4": {
			arg: `one two three four five six seven eight nine ten eleven twelve thirteen fourteen fifteen`,
			want: `
one two three four
five six seven eight
nine ten eleven twelve
thirteen fourteen fifteen`,
		},
		"t5": {
			arg: `one two three four five six seven eight nine ten eleven twelve thirteen fourteen fifteen sixteen`,
			want: `
one two three four
five six seven eight
nine ten eleven twelve
thirteen fourteen fifteen sixteen`,
		},
		"t6": {
			arg: `DTEng can get 25X VM 3DX Environments on-demand in < 5 biz days.`,
			want: `
DTEng can get 25X VM
3DX Environments on-demand in <
5 biz days.`,
		},
	}
	for n, tc := range tests {
		t.Run(n, func(t *testing.T) {
			w := tc.want[1:]
			if got := ShortLines(tc.arg); got != w {
				t.Errorf("ShortLines() = %v, want %v", got, w)
			}
		})
	}
}
