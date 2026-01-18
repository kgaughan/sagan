package model

import "time"

// Helper represents a set of command executed to do things such as manage a
// tunnel, fetch credentials, &c., needed by the workflows.
type Helper struct {
	Type     string        `yaml:"type"`
	Requires []string      `yaml:"requires,omitempty"`
	Args     []Argument    `yaml:"args,omitempty"`
	Commands []Command     `yaml:"run"`
	Duration time.Duration `yaml:"ttl,omitempty"`
}
