package model

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/kgaughan/sagan/internal/logging"
	"github.com/kgaughan/sagan/internal/utils"
)

// Task represents something on which a workflow operates.
//
// It can be dependent on another task having run and runs of this task can
// trigger other tasks to be implicitly re-executed.
type Task struct {
	Path       string    `yaml:"path"`
	Name       string    `yaml:"name"`
	Workflow   string    `yaml:"workflow"`
	Requires   []string  `yaml:"requires,omitempty"`
	Helpers    []string  `yaml:"helpers,omitempty"`
	Outputs    []Output  `yaml:"outputs,omitempty"`
	RedeployOn []Trigger `yaml:"redeploy_on,omitempty"`
}

func (t *Task) Normalize() {
	if t.Workflow == "" {
		t.Workflow = "default"
	}
	// extract a name from the path if none is specified
	if t.Name == "" {
		base := filepath.Base(t.Path)
		if base != "." && base != "/" && base != "" {
			t.Name = base
		}
	}
}

// Execute runs the workflow for a single task. It runs stage `Run` commands
// in topological order (based on stage requires) and executes `Finalize`
// commands in the reverse order. If a command has `SaveAs` set, the stdout is
// saved into an environment variable with that name for subsequent commands.
func (t Task) Execute(ctx context.Context, workflows map[string]*Workflow, dryRun bool, env map[string]string, envMu *sync.Mutex, logCh chan<- logging.TaskLog) error {
	wf, ok := workflows[t.Workflow]
	if !ok {
		return fmt.Errorf("workflow %q not found", t.Workflow)
	}

	// build stage graph: dependency -> dependents
	stageGraph := map[string][]string{}
	// ensure all stages are present
	for name := range wf.Stages {
		stageGraph[name] = []string{}
	}
	for name, st := range wf.Stages {
		for _, depStage := range st.Requires {
			// depStage is the name of a required stage
			if _, ok := stageGraph[depStage]; !ok {
				stageGraph[depStage] = []string{}
			}
			stageGraph[depStage] = append(stageGraph[depStage], name)
		}
	}

	order, err := utils.TopologicalSort(stageGraph)
	if err != nil {
		return fmt.Errorf("could not sort stages for task %v: %w", t.Path, err)
	}

	finalizers := []struct {
		stage string
		cmds  []Command
	}{}

	for _, stageName := range order {
		stage := wf.Stages[stageName]
		for _, cmd := range stage.Run {
			if err := cmd.Run(ctx, t.Path, dryRun, env, envMu, logCh, t.Name); err != nil {
				return fmt.Errorf("task %v stage %v run failed: %w", t.Path, stageName, err)
			}
		}
		if len(stage.Finalize) > 0 {
			finalizers = append(finalizers, struct {
				stage string
				cmds  []Command
			}{stage: stageName, cmds: stage.Finalize})
		}
	}

	// run finalizers in reverse order
	for i := len(finalizers) - 1; i >= 0; i-- {
		f := finalizers[i]
		for _, cmd := range f.cmds {
			if err := cmd.Run(ctx, t.Path, dryRun, env, envMu, logCh, t.Name); err != nil {
				return fmt.Errorf("task %v stage %v finalize failed: %w", t.Path, f.stage, err)
			}
		}
	}

	return nil
}
