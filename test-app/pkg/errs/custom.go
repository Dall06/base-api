package errs

type customError struct {
	sentinel error
	msg      string
}

func (e *customError) Error() string {
	return e.msg
}

func (e *customError) Unwrap() error {
	return e.sentinel
}
