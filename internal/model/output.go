package model

// Output represents something to be written to a configuration file upon the
// completion of a task run. Changes in values may trigger the implicit
// re-execution of tasks.
type Output struct {
	Path   string `yaml:"path"`
	Action string `yaml:"action"`
	Field  string `yaml:"field,omitempty"`
}
