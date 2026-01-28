package agent

type MisconfigurationError struct {
	err error
}

func NewMisconfigurationError(err error) error {
	return &MisconfigurationError{err: err}
}

func (e *MisconfigurationError) Error() string {
	return e.err.Error()
}

func (e *MisconfigurationError) Unwrap() error {
	return e.err
}
