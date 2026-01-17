package config

import (
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the root configuration object.
type Config struct {
	Version   string              `yaml:"version"`
	Helpers   map[string]Helper   `yaml:"helpers,omitempty"`
	Workflows map[string]Workflow `yaml:"workflows"`
	Projects  []Project           `yaml:"projects"`
}

// Helper represents a set of command executed to do things such as manage a
// tunnel, fetch credentials, &c., needed by the workflows.
type Helper struct {
	Type     string        `yaml:"type"`
	Requires []string      `yaml:"requires,omitempty"`
	Args     []Argument    `yaml:"args,omitempty"`
	Commands []Command     `yaml:"run"`
	Duration time.Duration `yaml:"ttl,omitempty"`
}

// Workflow represents a series of build stages with dependencies between them.
//
// Stages must specify their order via dependencies and will be sorted using a
// topological sort to figure out their execution and finalization order.
type Workflow struct {
	Temporaries []Temporary      `yaml:"temporaries,omitempty"`
	Sources     []string         `yaml:"load,omitempty"`
	Stages      map[string]Stage `yaml:",inline"`
}

// Project represents something on which a workflow operates.
//
// It can be dependent on another project having run and runs of this project
// can trigger other projects to be implicitly re-executed.
type Project struct {
	Path       string    `yaml:"path"`
	Name       string    `yaml:"name"`
	Workflow   string    `yaml:"workflow"`
	Requires   []string  `yaml:"requires,omitempty"`
	Helpers    []string  `yaml:"helpers,omitempty"`
	Outputs    []Output  `yaml:"outputs,omitempty"`
	RedeployOn []Trigger `yaml:"redeploy_on,omitempty"`
}

// Argument represents some value that a helper expects to be available.
type Argument struct {
	Name      string `yaml:"name"`
	Default   string `yaml:"default,omitempty"`
	Exclusive bool   `yaml:"exclusive"`
	Variable  string `yaml:"env,omitempty"`
}

// Command represents a command to be executed.
type Command struct {
	Command string `yaml:"cmd"`
	SaveAs  string `yaml:"save_as,omitempty"`
}

// Stage represents a series of commands to be executed followed by some
// commands to do cleanup afterwards.
type Stage struct {
	Requires map[string]string `yaml:"requires,omitempty"`
	Run      []Command         `yaml:"run,omitempty"`
	Finalize []Command         `yaml:"finalize,omitempty"`
}

// Temporary represents a temporary object of some kind, e.g., a file.
type Temporary struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

// Output represents something to be written to a configuration file upon the
// completion of a project run. Changes in values may trigger the implicit
// re-execution of projects.
type Output struct {
	Path   string `yaml:"path"`
	Action string `yaml:"action"`
	Field  string `yaml:"field,omitempty"`
}

// Trigger represents a configuration update that will lead to a project being
// re-executed.
type Trigger struct {
	Path  string `yaml:"path"`
	Field string `yaml:"field"`
}

// Load loads configuration from a YAML file at a given path.
func (c *Config) Load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open configuration: %w", err)
	}
	defer f.Close()

	if content, err := io.ReadAll(f); err != nil {
		return fmt.Errorf("could not read configuration: %w", err)
	} else if err := yaml.Unmarshal(content, c); err != nil {
		return fmt.Errorf("could not parse configuration: %w", err)
	}
	return nil
}
