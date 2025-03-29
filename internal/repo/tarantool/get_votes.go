package tarantool

import (
	"fmt"
	"vote-bot/internal/entity"

	"github.com/tarantool/go-tarantool/v2"
)

const (
	getVotesIndex = "vote_poll_id"
)

// GetVotes returns all votes that belong to poll with pollID.
// It is used to calculate poll results.
func (r *Repo) GetVotes(pollID uint64) ([]entity.Vote, error) {
	const op = "tarantool.repo.GetVotes"

	data, err := r.conn.Do(
		tarantool.NewSelectRequest(voteSpace).
			Index(getVotesIndex).
			Key([]any{int(pollID)}),
	).Get()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get votes: %w", op, err)
	}

	return serializeVotes(data), nil
}

// Converts tarantool response of type []any to slice []entity.Vote
func serializeVotes(tuples []any) []entity.Vote {
	const op = "repo.tarantool.serializeVotes"

	fmt.Println(tuples, len(tuples))
	votes := make([]entity.Vote, 0, len(tuples))

	for _, el := range tuples {
		tuple := el.([]any)

		// convert option nums to uint64 slice
		optsAny := tuple[3].([]any)
		opts := make([]uint64, 0, len(optsAny))
		for _, optAny := range optsAny {
			opts = append(opts, uint64(optAny.(int8)))
		}

		votes = append(votes, entity.Vote{
			VoteID:    uint64(tuple[0].(int8)),
			User:      tuple[1].(string),
			PollID:    uint64(tuple[2].(int8)),
			OptionIDs: opts,
		})
	}

	return votes
}
