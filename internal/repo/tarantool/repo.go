package tarantool

import (
	"github.com/tarantool/go-tarantool/v2"
)

const (
	pollSpace   = "polls"
	optionSpace = "options"
)

// Repo wraps a Tarantool connection to abstract database interactions.
type Repo struct {
	conn *tarantool.Connection
}

func NewRepo(conn *tarantool.Connection) *Repo {
	return &Repo{conn: conn}
}
