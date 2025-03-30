package repo

import "errors"

var (
	ErrPollDoesNotExist = errors.New("poll with this id does not exist")
	ErrNoOptionsFound   = errors.New("no options for the poll was found")
)
