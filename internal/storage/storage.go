package storage

import "errors"

var (
	ErrUrlExist    = errors.New("alias exist")
	ErrUrlNotFound = errors.New("alias not found")
)
