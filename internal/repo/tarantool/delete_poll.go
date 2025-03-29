package tarantool

import (
	"fmt"

	"github.com/tarantool/go-tarantool/v2"
)

const (
	deleteVotesFunc   = "delete_votes"
	deleteOptionsFunc = "delete_options"
)

// DeletePoll deletes the whole information about poll from spaces: votes, options, polls.
// It uses lua-defined functions delete_votes() and delete_options() under the hood.
func (r *Repo) DeletePoll(pollID uint64) error {
	const op = "repo.tarantool.DeletePoll"

	// Starts stream.
	stream, err := r.conn.NewStream()
	if err != nil {
		return fmt.Errorf("%s: failed to init stream for txn: %w", op, err)
	}

	// Delete votes using delete_votes() func.
	_, err = stream.Do(
		tarantool.NewCall17Request(deleteVotesFunc).
			Args([]any{pollID}),
	).Get()
	if err != nil {
		_, er := stream.Do(
			tarantool.NewRollbackRequest(),
		).Get()
		if er != nil {
			return fmt.Errorf("%s: failed to roll back DeletePoll txn: %w", op, er)
		}

		return fmt.Errorf("%s: failed to delete votes: %w", op, err)
	}

	// Delete options using delete_options() func.
	_, err = stream.Do(
		tarantool.NewCall17Request(deleteOptionsFunc).
			Args([]any{pollID}),
	).Get()
	if err != nil {
		_, er := stream.Do(
			tarantool.NewRollbackRequest(),
		).Get()
		if er != nil {
			return fmt.Errorf("%s: failed to roll back DeletePoll txn: %w", op, er)
		}

		return fmt.Errorf("%s: failed to delete options: %w", op, err)
	}

	// Finally, delete poll itself.
	_, err = stream.Do(
		tarantool.NewDeleteRequest(pollSpace).
			Key([]any{uint(pollID)}),
	).Get()
	if err != nil {
		_, er := stream.Do(
			tarantool.NewRollbackRequest(),
		).Get()
		if er != nil {
			return fmt.Errorf("%s: failed to roll back DeletePoll txn: %w", op, er)
		}

		return fmt.Errorf("%s: failed to delete poll: %w", op, err)
	}

	// Close stream (commit txn).
	_, err = stream.Do(
		tarantool.NewCommitRequest(),
	).Get()
	if err != nil {
		return fmt.Errorf("%s: failed to commit txn: %w", op, err)
	}

	return nil
}
