package entity

type Poll struct {
	ID          uint64
	Name        string
	Creator     string
	Channel     string
	IsFinished  bool
	IsMultiVote bool
}
