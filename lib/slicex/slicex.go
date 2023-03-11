package slicex

func Map[S any, D any](in []S, f func(S) D) []D {
	res := make([]D, 0, len(in))
	for _, i := range in {
		res = append(res, f(i))
	}
	return res
}
