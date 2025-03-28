package tarantool

import (
	"fmt"
	"vote-bot/internal/entity"
	"vote-bot/internal/repo"

	"github.com/tarantool/go-tarantool/v2"
)

const (
	pollSpace = "polls"
)

// Repo wraps a Tarantool connection to abstract database interactions.
type Repo struct {
	conn *tarantool.Connection
}

func NewRepo(conn *tarantool.Connection) *Repo {
	return &Repo{conn: conn}
}

func (r *Repo) createPoll(poll entity.Poll) (*entity.Poll, error) {
	const op = "repo.tarantool.createPoll"

	// Plural form because tarantool query returns slice of tuples
	// but only the first one is needed.
	var newPolls []entity.Poll

	tuple := []any{nil, poll.Name, poll.Creator, poll.Channel, nil, poll.IsMultiVote}
	err := r.conn.Do(
		tarantool.NewInsertRequest(pollSpace).
			Tuple(tuple),
	).GetTyped(&newPolls)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to insert poll: %w", op, err)
	}

	return &newPolls[0], nil
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
