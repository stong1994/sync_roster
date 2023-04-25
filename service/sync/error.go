package sync

import "errors"

var (
	ErrDeptNameOccupied = errors.New("dept name occupied")
)

var (
	errNotFound = errors.New("not found user")
	IDOccupied  = errors.New("id occupied")
)
