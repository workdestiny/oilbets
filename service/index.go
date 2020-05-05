package service

import (
	"time"
)

// Config is type for service config
type Config struct {
	Location *time.Location
}

var (
	loc *time.Location
)

// New is init config for service
func New(c *Config) {
	loc = c.Location
}
