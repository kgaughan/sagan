package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/kgaughan/sagan/internal/common"
	"github.com/kgaughan/sagan/internal/config"
	"github.com/kgaughan/sagan/internal/logging"
	"github.com/kgaughan/sagan/internal/orchestration"
	"github.com/kgaughan/sagan/internal/toposort"
	"github.com/kgaughan/sagan/internal/version"
	flag "github.com/spf13/pflag"
)

func main() {
	flag.Parse()

	if *PrintVersion {
		fmt.Println(version.Version)
		return
	}
	if *ShowHelp {
		flag.Usage()
		os.Exit(0)
	}

	cfg := &config.Config{}
	if err := cfg.Load(*ConfigPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	graph, tasks := cfg.BuildDependencyGraph()
	if len(tasks) == 0 {
		fmt.Fprintln(os.Stderr, "nothing to execute")
		os.Exit(0)
	}

	// linearize to check for cycles
	_, err := toposort.TopologicalSort(graph)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx := context.Background()

	logCh := make(chan logging.TaskLog, 512)
	go func() {
		for {
			log, ok := <-logCh
			if !ok {
				break
			}
			fmt.Printf("%v: %v\n", log.Task, log.Line)
		}
	}()

	sched := orchestration.NewScheduler(graph)

	statuses := map[string]string{}
	for k := range tasks {
		statuses[k] = "waiting"
	}

	env := map[string]string{}
	var envMu sync.Mutex
	var statusMu sync.Mutex

	_, err = sched.Run(ctx, *Workers, func(name string) error {
		t, ok := tasks[name]
		if !ok {
			return fmt.Errorf("%v: %w", name, common.ErrUnknownTask)
		}

		// update UI: mark task running
		statusMu.Lock()
		statuses[name] = "running"
		statusMu.Unlock()

		// run the task
		if err := t.Execute(ctx, cfg.Workflows, *DryRun, env, &envMu, logCh); err != nil {
			statusMu.Lock()
			statuses[name] = "failed"
			statusMu.Unlock()
			return err // nolint:wrapcheck
		}

		statusMu.Lock()
		statuses[name] = "done"
		statusMu.Unlock()

		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("final status:")
	for task, status := range statuses {
		fmt.Printf("  %v: %v\n", task, status)
	}
}
