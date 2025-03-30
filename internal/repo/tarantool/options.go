package tarantool

import (
	"fmt"
	"vote-bot/internal/entity"
	"vote-bot/internal/repo"

	"github.com/tarantool/go-tarantool/v2"
)

const (
	optionPollIndex = "option_poll_id"
)

// GetOptions returns all options that belong to poll with pollID.
func (r *Repo) GetOptions(pollID uint64) ([]entity.Option, error) {
	const op = "repo.tarantool.GetOptions"

	var options []entity.Option

	err := r.conn.Do(
		tarantool.NewSelectRequest(optionSpace).
			Index(optionPollIndex).
			Key([]any{int(pollID)}),
	).GetTyped(&options)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get options: %w", op, err)
	}

	if len(options) == 0 {
		return nil, fmt.Errorf("%s: %w", op, repo.ErrNoOptionsFound)
	}

	return options, nil
}
