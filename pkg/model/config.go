package model

import "github.com/peterbourgon/ff/v4"

type Config struct {
	*ff.Command
	*ff.FlagSet
}

func (c Config) IsZero() bool { return c.Command == nil && c.FlagSet == nil }
