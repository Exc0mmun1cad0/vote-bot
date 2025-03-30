package service

import "errors"

var (
	ErrPollNotFound = errors.New("poll with this id not found")
	ErrNotPollOwner = errors.New("user is not the owner of the poll")
	ErrPollFinished = errors.New("poll was finished")

	ErrNoVoteToCancel = errors.New("no vote to cancel")
	ErrNoVotesInPoll = errors.New("no votes in poll yet")

	ErrOnlyOneOptionAllowed = errors.New("only one option in the poll is allowed")
	ErrInvalidOptionNumber  = errors.New("there is option with invalid number")
)
