package utility

import "errors"

var ErrNotFound = errors.New("not found")
var ErrInvalidArgument = errors.New("invalid argument")
var ErrTimedOut = errors.New("timed out")
var ErrAlreadyExists = errors.New("already exists")
var ErrInvalidState = errors.New("invalid state")
var ErrSmartContractError = errors.New("smart contract error")
var ErrDatabase = errors.New("database error")
