package service

import (
	"errors"
	"fmt"
	"vote-bot/internal/entity"
	"vote-bot/internal/repo"
)

type PollRepo interface {
	CreatePollWithOptions(poll entity.Poll, options []entity.Option) (*entity.Poll, []entity.Option, error)
	GetPoll(pollID uint64) (*entity.Poll, error)
	FinishPoll(pollID uint64) error
	DeletePoll(pollID uint64) error
}

type PollService struct {
	pollRepo PollRepo
}

func NewPollService(pollRepo PollRepo) *PollService {
	return &PollService{pollRepo: pollRepo}
}

func (s *PollService) CreatePoll(poll entity.Poll, options []entity.Option) (*entity.Poll, []entity.Option, error) {
	const op = "service.CreatePoll"

	newPoll, newOptions, err := s.pollRepo.CreatePollWithOptions(poll, options)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: failed to create poll: %w", op, err)
	}

	return newPoll, newOptions, err
}

func (s *PollService) FinishPoll(pollID uint64, user string, channel string) error {
	const op = "service.FinishPoll"

	poll, err := s.pollRepo.GetPoll(pollID)
	if err != nil {
		if errors.Is(err, repo.ErrPollDoesNotExist) {
			return fmt.Errorf("%s: %w", op, ErrPollNotFound)
		}

		return fmt.Errorf("%s: failed to get poll: %w", op, err)
	}

	if poll.Channel != channel {
		return fmt.Errorf("%s: %w", op, ErrPollNotFound)
	}

	if poll.Creator != user {
		return fmt.Errorf("%s: %w", op, ErrNotPollOwner)
	}

	return s.pollRepo.FinishPoll(pollID)
}

func (s *PollService) DeletePoll(pollID uint64, user string, channel string) error {
	const op = "service.DeletePoll"

	poll, err := s.pollRepo.GetPoll(pollID)
	if err != nil {
		if errors.Is(err, repo.ErrPollDoesNotExist) {
			return fmt.Errorf("%s: %w", op, ErrPollNotFound)
		}

		return fmt.Errorf("%s: failed to get poll: %w", op, err)
	}

	if poll.Channel != channel {
		return fmt.Errorf("%s: %w", op, ErrPollNotFound)
	}

	if poll.Creator != user {
		return fmt.Errorf("%s: %w", op, ErrNotPollOwner)
	}

	return s.pollRepo.DeletePoll(pollID)
}
