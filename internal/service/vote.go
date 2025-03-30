package service

import (
	"errors"
	"fmt"
	"vote-bot/internal/entity"
	"vote-bot/internal/repo"
)

type VoteRepo interface {
	// for checking
	GetPoll(pollID uint64) (*entity.Poll, error)
	GetOptions(pollID uint64) ([]entity.Option, error)

	CreateVote(vote entity.Vote) (*entity.Vote, error)
	DeleteVote(user string, pollID uint64) (bool, error)
	GetVotes(pollID uint64) ([]entity.Vote, error)
}

type VoteService struct {
	voteRepo VoteRepo
}

func NewVoteService(voteRepo VoteRepo) *VoteService {
	return &VoteService{voteRepo: voteRepo}
}

func (s *VoteService) Vote(pollID uint64, user string, channel string, opts []uint64) error {
	const op = "service.Vote"

	poll, err := s.voteRepo.GetPoll(pollID)
	if err != nil {
		if errors.Is(err, repo.ErrPollDoesNotExist) {
			return fmt.Errorf("%s: %w", op, ErrPollNotFound)
		}
		return fmt.Errorf("%s: failed to find poll: %w", op, err)
	}

	if poll.Channel != channel {
		return fmt.Errorf("%s: %w", op, ErrPollNotFound)
	}

	if poll.IsFinished {
		return fmt.Errorf("%s: %w", op, ErrPollFinished)
	}

	var isManyOpts bool
	if len(opts) > 1 {
		isManyOpts = true
	}
	if isManyOpts && !poll.IsMultiVote {
		return fmt.Errorf("%s: %w", op, ErrOnlyOneOptionAllowed)
	}

	definedOptions, err := s.voteRepo.GetOptions(pollID)
	if err != nil {
		return fmt.Errorf("%s: failed to get options defined in the poll: %w", op, err)
	}
	for _, opt := range opts {
		if opt > uint64(len(definedOptions)) {
			return fmt.Errorf("%s: %w", op, ErrInvalidOptionNumber)
		}
	}

	_, err = s.voteRepo.CreateVote(entity.Vote{
		PollID:    pollID,
		OptionIDs: opts,
		User:      user,
	})
	if err != nil {
		return fmt.Errorf("%s: failed to create vote: %w", op, err)
	}

	return nil
}

func (s *VoteService) RetractVote(pollID uint64, user string, channel string) error {
	const op = "service.RetractVote"

	poll, err := s.voteRepo.GetPoll(pollID)
	if err != nil {
		if errors.Is(err, repo.ErrPollDoesNotExist) {
			return fmt.Errorf("%s: %w", op, ErrPollNotFound)
		}
		return fmt.Errorf("%s: failed to find poll: %w", op, err)
	}

	if poll.Channel != channel {
		return fmt.Errorf("%s: %w", op, ErrPollNotFound)
	}

	if poll.IsFinished {
		return fmt.Errorf("%s: %w", op, ErrPollFinished)
	}

	isDeleted, err := s.voteRepo.DeleteVote(user, pollID)
	if err != nil {
		return fmt.Errorf("%s: failed to delete vote: %w", op, err)
	}
	if !isDeleted {
		return fmt.Errorf("%s: %w", op, ErrNoVoteToCancel)
	}

	return nil
}

// GetResults returns poll resuts in the following format:
// map["1) optionName"] = optionCount.
func (s *VoteService) GetResults(pollID uint64, channel string) (map[string]uint64, error) {
	const op = "servoce.GetResults"

	// Get poll from repo to perform checks.
	poll, err := s.voteRepo.GetPoll(pollID)
	if err != nil {
		if errors.Is(err, repo.ErrPollDoesNotExist) {
			return nil, fmt.Errorf("%s: %w", op, ErrPollNotFound)
		}
		return nil, fmt.Errorf("%s: failed to find poll: %w", op, err)
	}

	// Check whether poll was created in the channel from which it is being requested.
	if poll.Channel != channel {
		return nil, fmt.Errorf("%s: %w", op, ErrPollNotFound)
	}

	// Get poll parameters that can be voted for.
	definedOptions, err := s.voteRepo.GetOptions(pollID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get options defined in the poll: %w", op, err)
	}

	// Create slice for saving statistics where
	// index is option number, and value is how many peopel voted for this option.
	results := make([]uint64, len(definedOptions))

	// Get votes in this poll. In case there are no votes return an error.
	votes, err := s.voteRepo.GetVotes(pollID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get votes: %w", op, err)
	}
	if len(votes) == 0 {
		return nil, fmt.Errorf("%s: %w", op, ErrNoVotesInPoll)
	}

	// Iterate over votes in order to form results 
	for _, vote := range votes {
		for _, opt := range vote.OptionIDs {
			results[opt-1]++
		}
	}

	// Format the results to make them human-readable.
	fmtResults := make(map[string]uint64, len(results))
	for option, count := range results {
		fmtResults[fmt.Sprintf("%d) %s", option+1, definedOptions[option].Name)] = count
	}

	return fmtResults, nil
}
