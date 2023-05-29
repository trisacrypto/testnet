package utils

import "errors"

var (
	ErrTooLong         = errors.New("base58 address too long")
	ErrInvalidChecksum = errors.New("invalid checksum")
)
