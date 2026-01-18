package model

// Workflow represents a series of build stages with dependencies between them.
//
// Stages must specify their order via dependencies and will be sorted using a
// topological sort to figure out their execution and finalization order.
type Workflow struct {
	Temporaries []Temporary      `yaml:"temporaries,omitempty"`
	Sources     []string         `yaml:"load,omitempty"`
	Stages      map[string]Stage `yaml:",inline"`
}
