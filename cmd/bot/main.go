package main

import (
	"log/slog"
	"os"
	"vote-bot/internal/bot"
	"vote-bot/internal/config"
	"vote-bot/pkg/sl"
	"vote-bot/pkg/tarantool"
)

func main() {
	cfg := config.MustLoad()

	log := bot.SetupLogger(cfg.Env)

	log.Debug("debug logs are enabled")

	log.Info(
		"starting bot...",
		slog.String("env", cfg.Env),
	)

	log.Info("initializing connection to tarantool...")
	conn, err := tarantool.NewConn(cfg.Tarantool)
	if err != nil {
		log.Error("failed to connect to db(postgres)", sl.Error(err))
		os.Exit(1)
	}
	log.Info("connected to tarantool")

	log.Info("initializing bot...")
	bot, err := bot.NewBot(log, cfg.Mattermost, conn)
	if err != nil {
		log.Error("failed to init bot", sl.Error(err))
		os.Exit(1)
	}

	// TODO: graceful shutdown
	bot.Client.ListenToEvents()
}
