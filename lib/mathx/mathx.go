package mathx

import "golang.org/x/exp/constraints"

func Max[T constraints.Ordered](v1, v2 T) T {
	if v1 > v2 {
		return v1
	}
	return v2
}
