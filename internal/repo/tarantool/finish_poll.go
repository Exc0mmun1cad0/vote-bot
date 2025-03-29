package tarantool

import (
	"fmt"

	"github.com/tarantool/go-tarantool/v2"
)

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
