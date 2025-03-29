package repo

import "errors"

var (
	ErrPollDoesNotExist = errors.New("poll with this id does not exist")
	ErrNoVotes          = errors.New("no votes found for poll with this id")
	ErrNoOptionsFound   = errors.New("no options for the poll was found")
)
