package client

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"time"
	"vote-bot/internal/config"
	"vote-bot/internal/service"
	"vote-bot/pkg/sl"

	"github.com/gorilla/websocket"
	"github.com/mattermost/mattermost-server/v6/model"
)

type Client struct {
	config  config.Mattermost
	l       *slog.Logger
	service *service.Service

	mattermostClient          *model.Client4
	mattermostWebSocketClient *model.WebSocketClient
	mattermostUser            *model.User
	mattermostTeam            *model.Team
}

func NewClient(cfg config.Mattermost, logger *slog.Logger, service *service.Service) (*Client, error) {
	const op = "bot.client.NewClient"

	client := &Client{
		config:  cfg,
		l:       logger,
		service: service,
	}

	log := client.l.With(slog.String("op", op))

	// Create a new mattermost client
	client.mattermostClient = model.NewAPIv4Client(cfg.Server.String())

	// Login
	client.mattermostClient.SetToken(cfg.Token)

	// Check authentication
	if user, resp, err := client.mattermostClient.GetUser("me", ""); err != nil {
		return nil, fmt.Errorf("%s: failed to log in: %w", op, err)
	} else {
		log.Debug("logged in", slog.Any("user", user), slog.Any("resp.StatusCode", resp.StatusCode))
		client.mattermostUser = user
	}

	// Find and save bot's team to app struct.
	if team, resp, err := client.mattermostClient.GetTeamByName(
		cfg.Team, "",
	); err != nil {
		return nil, fmt.Errorf("%s: could not find team. Is this bot a member ?: %w", op, err)
	} else {
		log.Debug("found team", slog.Any("team", team), slog.Any("resp.StatusCode", resp.StatusCode))
		client.mattermostTeam = team
	}

	return client, nil
}

// ListenToEvents connects to the Mattermost WebSocket API and listens for user events (messages).
func (c *Client) ListenToEvents() {
	const op = "bot.client.ListenToEvents"

	log := c.l.With(slog.String("op", op))

	var err error
	failCount := 0
	for {
		log.Debug("Establishing weebsocket connection to mattermost API...")
		c.mattermostWebSocketClient, err = model.NewWebSocketClient4WithDialer(
			&websocket.Dialer{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			fmt.Sprintf("ws://%s", c.config.Server.Host+c.config.Server.Path),
			c.mattermostClient.AuthToken,
		)
		if err != nil {
			log.Warn("Mattermost websocket disconnected, retrying", sl.Error(err))
			failCount += 1
			if failCount == 5 {
				time.Sleep(10 * time.Second)
			}
			continue
		}
		log.Debug("Established connection")
		log.Info("Mattermost websocket connected")

		c.mattermostWebSocketClient.Listen()

		for event := range c.mattermostWebSocketClient.EventChannel {
			c.handleEvent(event)
		}
	}
}

func (c *Client) StopListening() {
	const op = "bot.client.StopListening"

	c.l.Info("closing web socket connection...", slog.String("op", op))

	c.mattermostWebSocketClient.Close()
}
