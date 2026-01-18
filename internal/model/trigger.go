package model

// Trigger represents a configuration update that will lead to a task being
// re-executed.
type Trigger struct {
	Path  string `yaml:"path"`
	Field string `yaml:"field"`
}
