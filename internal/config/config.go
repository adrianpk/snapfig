package config

import (
	"errors"
	"strings"
)

type Config struct {
	Git string
	Path string
}

func (c *Config) Validate() error {
	git := strings.ToLower(c.Git)
	if git != "remove" && git != "disable" {
		return errors.New("'git' can only be 'remove' or 'disable'")
	}
	return nil
}
