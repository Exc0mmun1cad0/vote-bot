package entity

type Vote struct {
	PollID   uint64
	OptionID uint64
	User     string
}