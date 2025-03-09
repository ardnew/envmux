package config

import (
	"github.com/ardnew/envmux/config/parse"
)

type Parser struct {
	parse.Config
}

func MakeParser() Parser {
	return Parser{Config: parse.Make(parse.ConfigLexer)}
}
