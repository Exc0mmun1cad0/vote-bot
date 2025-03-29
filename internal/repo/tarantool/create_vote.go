package tarantool

import (
	"fmt"
	"vote-bot/internal/entity"

	"github.com/tarantool/go-tarantool/v2"
)

const (
	createVoteFunc = "create_vote"
)

// CreateVote adds new vote record to space "votes".
// It uses lua-defined create_vote() func under the hood.
//
// In case user votes second time, his vote will simply be updated.
func (r *Repo) CreateVote(vote entity.Vote) (*entity.Vote, error) {
	const op = "repo.tarantool.CreateVote"

	data, err := r.conn.Do(
		tarantool.NewCall17Request(createVoteFunc).
			Args([]any{vote.User, vote.PollID, vote.OptionIDs}),
	).Get()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create vote: %w", op, err)
	}

	tuple := data[0].([]any)
	tupleIDs := tuple[3].([]any)
	optionIDs := make([]uint64, len(tupleIDs))
	for i := range optionIDs {
		optionIDs[i] = uint64(tupleIDs[i].(int8))
	}
	newVote := entity.Vote{
		VoteID:    uint64(tuple[0].(int8)),
		User:      tuple[1].(string),
		PollID:    uint64(tuple[2].(int8)),
		OptionIDs: optionIDs,
	}

	return &newVote, nil
}
