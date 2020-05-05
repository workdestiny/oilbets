package app

type appError struct {
	value string
}

func (err *appError) Error() string {
	return err.value
}

// IsAppError checks is an error app error
func IsAppError(err error) bool {
	_, ok := err.(*appError)
	return ok
}

// NewAppError creates new app's error
func NewAppError(value string) error {
	return &appError{value}
}
