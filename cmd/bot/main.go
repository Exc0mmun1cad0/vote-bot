package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
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

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	log.Info("starting bot...")
	go bot.Client.ListenToEvents()

	<-stop
	log.Info("stopping app")

	bot.Client.StopListening()
	log.Info("bot doesn't listening for events anymore")

	tarantool.CloseConn(conn)
	log.Info("closed connection to tarantool")

	log.Info("stopped bot")

}
