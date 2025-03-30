package bot

import (
	"fmt"
	"log/slog"
	"vote-bot/internal/bot/client"
	"vote-bot/internal/config"
	tarantoolrepo "vote-bot/internal/repo/tarantool"
	"vote-bot/internal/service"

	"github.com/tarantool/go-tarantool/v2"
)

type Bot struct {
	log    *slog.Logger
	Client *client.Client
}

// NewBot initializes a new Mattermost bot instance.
func NewBot(log *slog.Logger, cfg config.Mattermost, conn *tarantool.Connection) (*Bot, error) {
	const op = "Bot.New"

	repo := tarantoolrepo.NewRepo(conn)
	service := service.NewService(repo)

	client, err := client.NewClient(cfg, log, service)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to initialize mattermost bot: %w", op, err)
	}

	return &Bot{
		log:    log,
		Client: client,
	}, nil
}
