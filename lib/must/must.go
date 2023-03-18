package must

import (
	"fmt"
	"os"
)

func Succeed(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func SucceedVal[T any](v T, err error) T {
	Succeed(err)
	return v
}
