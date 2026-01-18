package model

import (
	"bytes"
	"context"
	"maps"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/kgaughan/sagan/internal/logging"
)

// Command represents a command to be executed.
type Command struct {
	Command string `yaml:"cmd"`
	SaveAs  string `yaml:"save_as,omitempty"`
}

// Run executes a single command string through the shell. If Command.SaveAs
// is set, stdout is captured and stored in an environment variable with that
// name for subsequent commands.
func (c Command) Run(ctx context.Context, workdir string, dryRun bool, env map[string]string, envMu *sync.Mutex, logCh chan<- logging.TaskLog, taskName string) error {
	shell := "sh"
	arg := "-c"
	if dryRun {
		// do not execute, but mimic SaveAs by setting empty value
		if c.SaveAs != "" {
			envMu.Lock()
			env[c.SaveAs] = ""
			envMu.Unlock()
		}
		return nil
	}
	cmd := exec.CommandContext(ctx, shell, arg, c.Command)
	if workdir != "" {
		cmd.Dir = workdir
	}

	// inherit environment from parent process, then overlay env map
	baseEnv := os.Environ()
	// create a map to track overrides
	envMap := map[string]string{}
	for _, e := range baseEnv {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}
	maps.Copy(envMap, env)
	finalEnv := []string{}
	for k, v := range envMap {
		finalEnv = append(finalEnv, k+"="+v)
	}
	cmd.Env = finalEnv

	// Prepare pipes to stream output
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	var capture bytes.Buffer

	if err := cmd.Start(); err != nil {
		return err
	}

	// stream stdout
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdoutPipe.Read(buf)
			if n > 0 {
				chunk := string(buf[:n])
				capture.WriteString(chunk)
				// send each line if logCh is present
				if logCh != nil {
					lines := strings.Split(chunk, "\n")
					for i, l := range lines {
						if i == len(lines)-1 && l == "" {
							continue
						}
						logCh <- logging.TaskLog{Task: taskName, Line: l}
					}
				} else {
					os.Stdout.Write([]byte(chunk))
				}
			}
			if err != nil {
				break
			}
		}
	}()

	// stream stderr
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderrPipe.Read(buf)
			if n > 0 {
				chunk := string(buf[:n])
				capture.WriteString(chunk)
				if logCh != nil {
					lines := strings.Split(chunk, "\n")
					for i, l := range lines {
						if i == len(lines)-1 && l == "" {
							continue
						}
						logCh <- logging.TaskLog{Task: taskName, Line: l}
					}
				} else {
					os.Stderr.Write([]byte(chunk))
				}
			}
			if err != nil {
				break
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		return err
	}

	if c.SaveAs != "" {
		val := strings.TrimSpace(capture.String())
		// persist in provided env map for subsequent commands
		envMu.Lock()
		env[c.SaveAs] = val
		envMu.Unlock()
	}

	return nil
}
