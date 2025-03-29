package entity

type Vote struct {
	VoteID    uint64
	PollID    uint64
	OptionIDs []uint64
	User      string
}
