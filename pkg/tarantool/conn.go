package tarantool

import (
	"context"
	"fmt"
	"time"
	"vote-bot/internal/config"

	"github.com/tarantool/go-tarantool/v2"
)

func NewConn(cfg config.Tarantool) (*tarantool.Connection, error) {
	const op = "pkg.tarantool.NewConn"

	// TODO: maybe move time to config?
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// TODO: should i add SSL?
	dialer := tarantool.NetDialer{
		Address:  fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		User:     cfg.User,
		Password: cfg.Password,
	}

	// TODO: maybe move to config too?
	opts := tarantool.Opts{
		Timeout: time.Second,
	}

	conn, err := tarantool.Connect(ctx, dialer, opts)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to connect to tarantool db: %w", op, err)
	}

	return conn, nil
}

func CloseConn(conn *tarantool.Connection) error {
	const op = "pkg.tarantool.MustClose"

	if err := conn.CloseGraceful(); err != nil {
		return fmt.Errorf("%s: failed to close tarantool connection: %w", op, err)
	}

	return nil
}
