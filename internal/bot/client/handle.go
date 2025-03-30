package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"vote-bot/internal/entity"
	"vote-bot/internal/service"
	"vote-bot/pkg/sl"

	"github.com/mattermost/mattermost-server/v6/model"
)

func (c *Client) handleEvent(event *model.WebSocketEvent) {
	const op = "bot.client.handle"

	log := c.l.With(
		slog.String("op", op),
	)

	log.Debug(
		"got new event",
		slog.String("type", event.EventType()),
	)

	// ignore event in case it's not post.
	if event.EventType() != model.WebsocketEventPosted {
		return
	}

	post := &model.Post{}

	log.Debug("unmarshalling post...")
	err := json.Unmarshal([]byte(event.GetData()["post"].(string)), &post)
	if err != nil {
		log.Error("failed to unmarshal post")
		return
	}

	// ignore messages sent by bot itself.
	if post.UserId == c.mattermostUser.Id {
		return
	}

	msg := post.Message
	log.Info(
		"got new message",
		slog.String("msg", msg),
		slog.String("channel_id", post.ChannelId),
		slog.String("user_id", post.UserId),
		slog.String("message_id", post.Id),
	)

	// skip non-command messages.
	if msg[0] != '!' {
		return
	}

	// Process command
	lines := strings.Split(strings.TrimSpace(msg), "\n")
	sep := strings.Index(lines[0], " ")
	if sep == -1 {
		c.sendMessage(post.ChannelId, "invalid command", post.Id)
		return
	}
	args := strings.Split(lines[0], " ")
	cmd, arg := args[0], strings.Join(args[1:], " ")

	switch cmd {
	case cmdCreatePoll, cmdCreateMultiPoll:
		var prefix string
		if cmd == cmdCreateMultiPoll {
			prefix = "multi"
		}

		poll := entity.Poll{
			Name:        arg,
			Creator:     post.UserId,
			Channel:     post.ChannelId,
			IsMultiVote: false,
		}

		if cmd == cmdCreateMultiPoll {
			poll.IsMultiVote = true
		}

		// Skip if no options specified
		if len(lines) <= 1 {
			log.Error(fmt.Sprintf("failed to create %spoll without options", prefix))
			c.sendMessage(post.ChannelId, fmt.Sprintf("%spoll without options cannot be created", prefix), post.Id)
			return
		}

		options := make([]entity.Option, 0, len(lines)-1)
		for _, line := range lines[1:] {
			options = append(options, entity.Option{Name: strings.TrimSpace(line)})
		}

		// Create (multi)poll with options
		newPoll, newOptions, err := c.service.PollService.CreatePoll(poll, options)
		if err != nil {
			log.Error(fmt.Sprintf("failed to create %spoll", prefix), sl.Error(err))
			c.sendMessage(post.ChannelId, fmt.Sprintf("failed to create %spoll", prefix), post.Id)
			return
		}

		// Form response
		var b strings.Builder
		b.WriteString(fmt.Sprintf("New %spoll created: %s\nID: %d\n", prefix, newPoll.Name, newPoll.ID))
		for _, opt := range newOptions {
			b.WriteString(fmt.Sprintf("%d) %s\n", opt.Num, opt.Name))
		}

		// Send response to channel
		c.sendMessage(post.ChannelId, b.String(), "")

		log.Info(fmt.Sprintf("created %spoll", prefix), slog.Any("poll", poll), slog.Any("options", options))

	case cmdFinishPoll:
		pollID, err := pollIDFromString(args[1])
		if err != nil {
			log.Error("invalid poll ID", sl.Error(err))
			c.sendMessage(post.ChannelId, "invalid poll ID", post.Id)
			return
		}

		err = c.service.PollService.FinishPoll(pollID, post.UserId, post.ChannelId)
		if err != nil {
			if errors.Is(err, service.ErrPollNotFound) {
				log.Error("poll not found", slog.Uint64("pollID", pollID))
				c.sendMessage(post.ChannelId, "poll not found", post.Id)
				return
			}

			if errors.Is(err, service.ErrNotPollOwner) {
				log.Error("failed to finish the poll: user is not the poll creator", slog.Uint64("pollID", pollID))
				c.sendMessage(post.ChannelId, "impossible to finish poll which you are not creator of", post.Id)
				return
			}

			log.Error("failed to finish poll", slog.Uint64("pollID", pollID), sl.Error(err))
			c.sendMessage(post.ChannelId, "failed to finish poll", post.Id)
			return
		}

		c.sendMessage(post.ChannelId, fmt.Sprintf("poll %d was finished", pollID), "")

		log.Info("poll was finished", slog.Uint64("poll_id", pollID))

	case cmdDeletePoll:
		pollID, err := pollIDFromString(args[1])
		if err != nil {
			log.Error("invalid poll ID", sl.Error(err))
			c.sendMessage(post.ChannelId, "invalid poll ID", post.Id)
			return
		}

		err = c.service.PollService.DeletePoll(pollID, post.UserId, post.ChannelId)
		if err != nil {
			if errors.Is(err, service.ErrPollNotFound) {
				log.Error("poll not found", slog.Uint64("pollID", pollID))
				c.sendMessage(post.ChannelId, "poll not found", post.Id)
				return
			}

			if errors.Is(err, service.ErrNotPollOwner) {
				log.Error("failed to delete the poll: user is not the poll creator", slog.Uint64("pollID", pollID))
				c.sendMessage(post.ChannelId, "impossible to delete poll which you are not creator of", post.Id)
				return
			}

			log.Error("failed to delete poll", slog.Uint64("pollID", pollID), sl.Error(err))
			c.sendMessage(post.ChannelId, "failed to delete poll", post.Id)
			return
		}

		c.sendMessage(post.ChannelId, fmt.Sprintf("poll %d was deleted", pollID), "")

		log.Info("poll was deleted", slog.Uint64("poll_id", pollID))

	case cmdVote:
		pollID, err := pollIDFromString(args[1])
		if err != nil {
			log.Error("invalid poll ID", sl.Error(err))
			c.sendMessage(post.ChannelId, "invalid poll ID", post.Id)
			return
		}

		opts, err := optsFromString(lines[1])
		if err != nil {
			log.Error("invalid option nums", slog.Any("options", opts), sl.Error(err))
			c.sendMessage(post.ChannelId, "invalid options", post.Id)
			return
		}

		err = c.service.VoteService.Vote(pollID, post.UserId, post.ChannelId, opts)
		if err != nil {
			if errors.Is(err, service.ErrPollNotFound) {
				log.Error("poll not found", slog.Uint64("pollID", pollID))
				c.sendMessage(post.ChannelId, "poll not found", post.Id)
				return
			}

			if errors.Is(err, service.ErrPollFinished) {
				log.Error("failed to vote because poll was finished", slog.Uint64("pollID", pollID))
				c.sendMessage(post.ChannelId, "failed to vote because poll was finished", post.Id)
				return
			}

			if errors.Is(err, service.ErrOnlyOneOptionAllowed) {
				log.Error("failed to vote because it doesn't support multiple options", slog.Uint64("pollID", pollID))
				c.sendMessage(post.ChannelId, "failed to vote because it doesn't support multiple options", post.Id)
				return
			}

			if errors.Is(err, service.ErrInvalidOptionNumber) {
				log.Error("invalid option number", slog.Uint64("pollID", pollID), slog.Any("options", opts))
				c.sendMessage(post.ChannelId, "invalid option number", post.Id)
				return
			}

			log.Error("failed to vote", slog.Uint64("pollID", pollID), slog.Any("options", opts), sl.Error(err))
			c.sendMessage(post.ChannelId, "failed to vote", post.Id)
			return
		}

		c.sendMessage(post.ChannelId, "your vote was counted", "")
		log.Info("user voted", slog.Uint64("poll_id", pollID), slog.String("userId", post.UserId))

	case cmdRetractVote:
		pollID, err := pollIDFromString(args[1])
		if err != nil {
			log.Error("invalid poll ID", sl.Error(err))
			c.sendMessage(post.ChannelId, "invalid poll ID", post.Id)
			return
		}

		err = c.service.VoteService.RetractVote(pollID, post.UserId, post.ChannelId)
		if err != nil {
			if errors.Is(err, service.ErrPollNotFound) {
				log.Error("poll not found", slog.Uint64("pollID", pollID))
				c.sendMessage(post.ChannelId, "poll not found", post.Id)
				return
			}

			if errors.Is(err, service.ErrPollFinished) {
				log.Error("poll was finished", slog.Uint64("pollID", pollID))
				c.sendMessage(post.ChannelId, "poll was finished", post.Id)
				return
			}

			if errors.Is(err, service.ErrNoVoteToCancel) {
				log.Error("no vote to cancel", slog.Uint64("pollID", pollID))
				c.sendMessage(post.ChannelId, "you haven't vote yet in this poll", post.Id)
				return
			}

			log.Error("failed to retract vote", slog.Uint64("pollID", pollID), sl.Error(err))
			c.sendMessage(post.ChannelId, "failed to retract vote", post.Id)
			return
		}

		c.sendMessage(post.ChannelId, "your vote was retracted", "")
		log.Info("user retracted vote", slog.Uint64("poll_id", pollID), slog.String("userId", post.UserId))

	case cmdGetResults:
		pollID, err := pollIDFromString(args[1])
		if err != nil {
			log.Error("invalid poll ID", sl.Error(err))
			c.sendMessage(post.ChannelId, "invalid poll ID", post.Id)
			return
		}

		results, err := c.service.VoteService.GetResults(pollID, post.ChannelId)
		if err != nil {
			if errors.Is(err, service.ErrPollNotFound) {
				log.Error("poll not found", slog.Uint64("pollID", pollID))
				c.sendMessage(post.ChannelId, "poll not found", "")
				return
			}

			if errors.Is(err, service.ErrNoVotesInPoll) {
				log.Error("no votes in poll", slog.Uint64("pollID", pollID))
				c.sendMessage(post.ChannelId, "no votes in poll", "")
				return
			}

			log.Error("failed to get poll results", slog.Uint64("pollID", pollID), sl.Error(err))
			c.sendMessage(post.ChannelId, "failed get poll results", "")
			return
		}

		var b strings.Builder // response
		b.WriteString("Results:\n")
		for key, val := range results {
			b.WriteString(fmt.Sprintf("%s: %d\n", key, val))
		}

		c.sendMessage(post.ChannelId, b.String(), "")
		log.Info("counted poll results", slog.Uint64("poll_id", pollID), slog.String("userId", post.UserId))
	}
}

func (c *Client) sendMessage(channel, message, replyToID string) {
	const op = "bot.client.sendMessage"

	log := c.l.With(slog.String("op", op))

	post := &model.Post{}

	post.ChannelId = channel
	post.Message = message
	post.RootId = replyToID

	if post, resp, err := c.mattermostClient.CreatePost(post); err != nil {
		log.Error("failed to send message", sl.Error(err))
	} else {
		log.Debug(
			"sended message",
			slog.String("text", message), slog.String("channel", post.ChannelId),
			slog.Int("resp", resp.StatusCode),
		)
	}
}

func pollIDFromString(pollIDStr string) (uint64, error) {
	const op = "bot.client.pollIDFromString"

	pollID, err := strconv.Atoi(pollIDStr)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return uint64(pollID), nil
}

func optsFromString(optionsStr string) ([]uint64, error) {
	const op = "bot.client.optsFromString"

	optsStr := strings.Split(optionsStr, " ")
	opts := make([]uint64, 0, len(optsStr))
	for _, optStr := range optsStr {
		opt, err := strconv.Atoi(optStr)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		opts = append(opts, uint64(opt))
	}

	return opts, nil
}
