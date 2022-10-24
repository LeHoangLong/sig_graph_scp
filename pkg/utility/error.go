package utility

import "errors"

var ErrNotFound = errors.New("not found")
var ErrInvalidArgument = errors.New("invalid argument")
var ErrTimedOut = errors.New("timed out")
var ErrAlreadyExists = errors.New("already exists")
