package model

// Stage represents a series of commands to be executed followed by some
// commands to do cleanup afterwards.
type Stage struct {
	Requires map[string]string `yaml:"requires,omitempty"`
	Run      []Command         `yaml:"run,omitempty"`
	Finalize []Command         `yaml:"finalize,omitempty"`
}
