package model

type Poll struct {
	ID          uint
	Creator     string
	Channel     string
	IsFinished  bool
	IsMultiVote bool
}

type Option struct {
	ID     uint
	PollID uint
	Name   string
}

type Vote struct {
	PollID   uint
	OptionID uint
	User     string
}
