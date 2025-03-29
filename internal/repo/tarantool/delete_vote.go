package tarantool

import (
	"fmt"

	"github.com/tarantool/go-tarantool/v2"
)

const (
	deleteVoteIndex = "vote_user_poll_id"
)

// Delete vote removes record from space "votes".
// It is called in order to retract vote.
//
// isDeleted indicates whether vote was deleted or not.
// (false, nil) indicates that user hasn't voted before.
func (r *Repo) DeleteVote(user string, pollID uint64) (isDeleted bool, err error) {
	const op = "repo.tarantool.DeleteVote"

	data, err := r.conn.Do(
		tarantool.NewDeleteRequest(voteSpace).
			Index(deleteVoteIndex).
			Key([]any{user, int(pollID)}),
	).Get()
	if err != nil {
		return false, fmt.Errorf("%s: failed to delete vote: %w", op, err)
	}

	if len(data) == 0 {
		return false, nil
	}

	return true, nil
}
