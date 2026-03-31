package mock

func CastOrDefault[T any](val any) (res T) {
	if val != nil {
		res = val.(T)
	}
	return
}

func ErrOrNil(val any) (err error) {
	return CastOrDefault[error](val)
}
