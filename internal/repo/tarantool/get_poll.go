package tarantool

import (
	"fmt"
	"vote-bot/internal/entity"
	"vote-bot/internal/repo"

	"github.com/tarantool/go-tarantool/v2"
)

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
