package constants

import "errors"

var (
	ErrCodeWrongEnumString = errors.New("err_wrong_enum_string")
	ErrCodeNotFound        = errors.New("err_not_found")
	ErrCodeMaxSizeExceeded = errors.New("err_max_size_exceeded")
	ErrCodeInvalidValue    = errors.New("err_invalid_value")
	ErrCodeUnknown         = errors.New("err_unknown")
	ErrShouldNotHappen     = errors.New("err_should_not_happen")
)
