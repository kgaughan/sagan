package config

import (
	"fmt"
	"io"
	"os"

	"github.com/kgaughan/sagan/internal/common"
	"github.com/kgaughan/sagan/internal/model"
	"gopkg.in/yaml.v3"
)

// Config represents the root configuration object.
type Config struct {
	Version   string                     `yaml:"version"`
	Helpers   map[string]*model.Helper   `yaml:"helpers,omitempty"`
	Workflows map[string]*model.Workflow `yaml:"workflows"`
	Tasks     []*model.Task              `yaml:"tasks"`
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

	c.normalize()
	return nil
}

func (c *Config) normalize() {
	for _, p := range c.Tasks {
		p.Normalize()
	}
}

// ValidateConfig performs basic sanity checks on the configuration.
// It verifies that every task references an existing workflow and that
// every declared requirement refers to a known task name (derived from
// the task's path).
func (c *Config) Validate() error {
	// workflows presence
	for _, p := range c.Tasks {
		if _, ok := c.Workflows[p.Workflow]; !ok {
			return fmt.Errorf("task %q references %q: %w", p.Path, p.Workflow, common.ErrUnknownWorkflow)
		}
	}

	// build name map for requires validation
	names := map[string]struct{}{}
	for _, t := range c.Tasks {
		names[t.Name] = struct{}{}
	}
	for i, t := range c.Tasks {
		if t.Name == "" {
			return fmt.Errorf("task #%v has no name", i+1) // nolint:err113
		}
		for _, req := range t.Requires {
			if _, ok := names[req]; !ok {
				return fmt.Errorf("task %q requires %q: %w", t.Path, req, common.ErrUnknownTask)
			}
		}
	}

	return nil
}

// BuildDependencyGraph constructs a graph suitable for TopologicalSort.
// The graph maps a node to the list of nodes that depend on it (edges
// are dependency -> dependent). It also returns a map of task names to
// tasks.
func (c Config) BuildDependencyGraph() (map[string][]string, map[string]*model.Task) {
	tasks := map[string]*model.Task{}
	for _, t := range c.Tasks {
		tasks[t.Name] = t
	}

	graph := map[string][]string{}
	// ensure every task appears in the graph
	for name := range tasks {
		graph[name] = []string{}
	}

	// add edges from requirement -> dependent
	for name, t := range tasks {
		for _, req := range t.Requires {
			// if the requirement isn't a known task, still create the node
			if _, ok := graph[req]; !ok {
				graph[req] = []string{}
			}
			graph[req] = append(graph[req], name)
		}
	}

	return graph, tasks
}
