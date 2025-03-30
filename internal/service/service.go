package service

type Repo interface {
	VoteRepo
	PollRepo
}

type Service struct {
	PollService *PollService
	VoteService *VoteService
}

func NewService(repo Repo) *Service {
	return &Service{
		PollService: NewPollService(repo),
		VoteService: NewVoteService(repo),
	}
}
