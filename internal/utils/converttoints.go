package utils

import (
	"fmt"
	"strconv"
)

func ConvertToInt(args []string) (n []int, err error) {
	n = make([]int, len(args))
	for i := 0; i < len(args); i++ {
		n[i], err = strconv.Atoi(args[i])
		if err != nil {
			return nil, fmt.Errorf("%q is not a number", args[i])
		}
	}
	return
}
