package model

type Poll struct {
	ID          uint64
	Creator     string
	Channel     string
	IsFinished  bool
	IsMultiVote bool
}
