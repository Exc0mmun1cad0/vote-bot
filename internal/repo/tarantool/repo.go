package tarantool

import (
	"fmt"
	"vote-bot/internal/entity"
	"vote-bot/internal/repo"

	"github.com/tarantool/go-tarantool/v2"
)

const (
	pollSpace   = "polls"
	optionSpace = "options"
)

// Repo wraps a Tarantool connection to abstract database interactions.
type Repo struct {
	conn *tarantool.Connection
}

func NewRepo(conn *tarantool.Connection) *Repo {
	return &Repo{conn: conn}
}

// CreatePollWithOprions inserts new poll and its options to pollSpace and optionSpace respectively.
//
// Uses createPoll and createOptions under the hood within stream (transaction).
func (r *Repo) CreatePollWithOptions(poll entity.Poll, options []entity.Option) (*entity.Poll, []entity.Option, error) {
	const op = "repo.tarantool.CreatePollWithOprions"

	// Create new stream for atomicity
	stream, err := r.conn.NewStream()
	if err != nil {
		return nil, nil, fmt.Errorf("%s: failed to init stream for txn: %w", op, err)
	}

	// Create poll within stream
	newPoll, err := CreatePoll(stream, poll)
	if err != nil {
		_, er := stream.Do(
			tarantool.NewRollbackRequest(),
		).Get()
		if er != nil {
			return nil, nil, fmt.Errorf("%s: failed to roll back createPoll txn: %w", op, er)
		}

		return nil, nil, fmt.Errorf("%s: failed to create poll: %w", op, err)
	}

	// Set PollID in options
	newPollID := newPoll.ID
	for i := range options {
		options[i].PollID = newPollID
	}

	// Create options within stream
	newOptions, err := CreateOptions(stream, options)
	if err != nil {
		_, er := stream.Do(
			tarantool.NewRollbackRequest(),
		).Get()
		if er != nil {
			return nil, nil, fmt.Errorf("%s: failed to roll back createOptions txn: %w", op, er)
		}

		return nil, nil, fmt.Errorf("%s: failed to create options: %w", op, err)
	}

	return newPoll, newOptions, nil
}

func CreatePoll(s *tarantool.Stream, poll entity.Poll) (*entity.Poll, error) {
	const op = "repo.tarantool.createPoll"

	// Plural form because tarantool query returns slice of tuples
	// but only the first one is needed.
	var newPolls []entity.Poll

	tuple := []any{nil, poll.Name, poll.Creator, poll.Channel, nil, poll.IsMultiVote}
	err := s.Do(
		tarantool.NewInsertRequest(pollSpace).
			Tuple(tuple),
	).GetTyped(&newPolls)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create poll: %w", op, err)
	}

	return &newPolls[0], nil
}

func CreateOptions(s *tarantool.Stream, options []entity.Option) ([]entity.Option, error) {
	const op = "repo.tarantool.createOptions"

	newOptions := make([]entity.Option, 0, len(options))

	var futures []*tarantool.Future
	for i, option := range options {
		tuple := []any{nil, option.PollID, option.Name, i + 1}
		request := tarantool.NewInsertRequest(optionSpace).Tuple(tuple)
		futures = append(futures, s.Do(request))
	}

	for _, future := range futures {
		var createdOptions []entity.Option

		err := future.GetTyped(&createdOptions)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to create options: %w", op, err)
		}

		newOptions = append(newOptions, createdOptions[0])
	}

	return newOptions, nil
}

// GetPoll retursn info about poll by its ID.
func (r *Repo) GetPoll(pollID uint64) (*entity.Poll, error) {
	const op = "repo.tarantool.GetPoll"

	// Plural form because tarantool query returns slice of tuples
	// but only the first one is needed.
	var polls []entity.Poll

	err := r.conn.Do(
		tarantool.NewSelectRequest(pollSpace).
			Key(tarantool.UintKey{I: uint(pollID)}),
	).GetTyped(&polls)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get poll by ID: %w", op, err)
	}
	if len(polls) == 0 {
		return nil, fmt.Errorf("%s: %w", op, repo.ErrPollDoesNotExist)
	}

	return &polls[0], nil
}

// FinishPoll sets field is_finished to true.
func (r *Repo) FinishPoll(pollID uint64) error {
	const op = "repo.tarantool.FinishPoll"

	_, err := r.conn.Do(
		tarantool.NewUpdateRequest(pollSpace).
			Key(tarantool.UintKey{I: uint(pollID)}).
			Operations(tarantool.NewOperations().Assign(4, true)),
	).Get()
	if err != nil {
		return fmt.Errorf("%s: failed to finish poll: %w", op, err)
	}

	return nil
}
