package model

// Temporary represents a temporary object of some kind, e.g., a file.
type Temporary struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}
