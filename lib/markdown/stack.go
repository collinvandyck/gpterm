package markdown

type stack[T any] []*T

func (s *stack[T]) push(v T) {
	*s = append(*s, &v)
}

func (s *stack[T]) pop() *T {
	if len(*s) == 0 {
		return nil
	}
	v := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return v
}

func (s *stack[T]) peek() *T {
	if len(*s) == 0 {
		return nil
	}
	return (*s)[len(*s)-1]
}
