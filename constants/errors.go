package constants

import "errors"

var (
	ErrWrongEnumString = errors.New("err_wrong_enum_string")
	ErrNotFound        = errors.New("err_not_found")
	ErrMaxSizeExceeded = errors.New("err_max_size_exceeded")
	ErrInvalidValue    = errors.New("err_invalid_value")
)
