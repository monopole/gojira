package utils

import "strings"

func ShortLines(x string) string {
	words := strings.Split(x, " ")
	wordsPerLine := 4
	if len(words)%wordsPerLine == 1 {
		// Avoid having 1 word on last line
		wordsPerLine = 3
		if len(words)%wordsPerLine == 1 {
			wordsPerLine = 5
		}
	}
	var b strings.Builder
	for i := 1; i < len(words); i++ {
		b.WriteString(words[i-1])
		if i%wordsPerLine == 0 {
			b.WriteByte('\n')
		} else {
			b.WriteByte(' ')
		}
	}
	if len(words) > 0 {
		b.WriteString(words[len(words)-1])
	}
	return b.String()
}
