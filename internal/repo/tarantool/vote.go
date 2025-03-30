package tarantool

import (
	"fmt"
	"vote-bot/internal/entity"

	"github.com/tarantool/go-tarantool/v2"
)

const (
	createVoteFunc  = "create_vote"
	deleteVoteIndex = "vote_user_poll_id"
	getVotesIndex   = "vote_poll_id"
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

// Converts tarantool response of type []any to []entity.Vote.
func serializeVotes(tuples []any) []entity.Vote {
	votes := make([]entity.Vote, 0, len(tuples))

	for _, el := range tuples {
		tuple := el.([]any)

		// convert option nums to uint64 slice.
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
