package model

// Argument represents some value that a helper expects to be available.
type Argument struct {
	Name      string `yaml:"name"`
	Default   string `yaml:"default,omitempty"`
	Exclusive bool   `yaml:"exclusive"`
	Variable  string `yaml:"env,omitempty"`
}
