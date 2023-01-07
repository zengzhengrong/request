package config

import "errors"

type StatusCodeErrorType error
type JsonKeyErrorType error

var (
	StatusCodeError = StatusCodeErrorType(errors.New("StatusCodeError"))
	JsonKeyError    = JsonKeyErrorType(errors.New("JsonKeyError"))
)
