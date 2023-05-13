package config

import (
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version   string              `yaml:"version"`
	Helpers   map[string]Helper   `yaml:"helpers,omitempty"`
	Workflows map[string]Workflow `yaml:"workflows"`
	Projects  []Project           `yaml:"projects"`
}

type Helper struct {
	Type     string        `yaml:"type"`
	Requires []string      `yaml:"requires,omitempty"`
	Args     []Argument    `yaml:"args,omitempty"`
	Commands []Command     `yaml:"run"`
	Duration time.Duration `yaml:"ttl,omitempty"`
}

type Workflow struct {
	Temporaries []Temporary      `yaml:"temporaries,omitempty"`
	Sources     []string         `yaml:"load,omitempty"`
	Stages      map[string]Stage `yaml:",inline"`
}

type Project struct {
	Path       string    `yaml:"path"`
	Workflow   string    `yaml:"workflow"`
	Requires   []string  `yaml:"requires,omitempty"`
	Helpers    []string  `yaml:"helpers,omitempty"`
	Outputs    []Output  `yaml:"outputs,omitempty"`
	RedeployOn []Trigger `yaml:"redeploy_on,omitempty"`
}

type Argument struct {
	Name      string `yaml:"name"`
	Default   string `yaml:"default,omitempty"`
	Exclusive bool   `yaml:"exclusive"`
	Variable  string `yaml:"env,omitempty"`
}

type Command struct {
	Command string `yaml:"cmd"`
	SaveAs  string `yaml:"save_as,omitempty"`
}

type Stage struct {
	Requires map[string]string `yaml:"requires,omitempty"`
	Run      []Command         `yaml:"run,omitempty"`
}

type Temporary struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type Output struct {
	Path   string `yaml:"path"`
	Action string `yaml:"action"`
	Field  string `yaml:"field,omitempty"`
}

type Trigger struct {
	Path  string `yaml:"path"`
	Field string `yaml:"field"`
}

func (c *Config) Load(path string) error {
	fh, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fh.Close()

	if content, err := io.ReadAll(fh); err != nil {
		return err
	} else if err := yaml.Unmarshal(content, c); err != nil {
		return err
	} else {
		return nil
	}
}
