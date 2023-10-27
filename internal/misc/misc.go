package misc

func PtrTo[T any](e T) *T {
	return &e
}
