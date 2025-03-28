package repo

import "errors"

var (
	ErrPollDoesNotExist = errors.New("Poll with this ID does not exist")
)
