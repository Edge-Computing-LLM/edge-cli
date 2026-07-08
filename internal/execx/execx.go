package execx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Runner struct {
	Verbose bool
	DryRun  bool
	Env     []string
}

type Command struct {
	Name    string
	Args    []string
	Dir     string
	Sudo    bool
	Mutates bool
}

func Exists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func (r Runner) Run(ctx context.Context, c Command) error {
	name, args := commandParts(c)
	if c.Mutates && r.DryRun {
		fmt.Printf("[dry-run] %s %s\n", name, strings.Join(args, " "))
		return nil
	}
	if r.Verbose {
		fmt.Printf("[run] %s %s\n", name, strings.Join(args, " "))
	}
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = c.Dir
	cmd.Env = append(os.Environ(), r.Env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (r Runner) Output(ctx context.Context, c Command) (string, error) {
	name, args := commandParts(c)
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = c.Dir
	cmd.Env = append(os.Environ(), r.Env...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return strings.TrimSpace(out.String()), err
}

func (r Runner) Check(ctx context.Context, label string, c Command) Result {
	out, err := r.Output(ctx, c)
	return Result{Label: label, Output: out, Err: err}
}

type Result struct {
	Label  string
	Output string
	Err    error
}

func (r Result) OK() bool { return r.Err == nil }

func PrintResults(results []Result) error {
	var failed []string
	for _, res := range results {
		if res.OK() {
			fmt.Printf("[ok] %s\n", res.Label)
		} else {
			fmt.Printf("[fail] %s\n", res.Label)
			failed = append(failed, res.Label)
		}
		if res.Output != "" {
			fmt.Println(indent(res.Output))
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("%d check(s) failed: %s", len(failed), strings.Join(failed, ", "))
	}
	return nil
}

func commandParts(c Command) (string, []string) {
	if c.Sudo && os.Geteuid() != 0 {
		return "sudo", append([]string{c.Name}, c.Args...)
	}
	return c.Name, c.Args
}

func RequiredCommands(names ...string) []Result {
	results := make([]Result, 0, len(names))
	for _, name := range names {
		if path, err := exec.LookPath(name); err == nil {
			results = append(results, Result{Label: "command " + name, Output: path})
		} else {
			results = append(results, Result{Label: "command " + name, Err: errors.New("missing")})
		}
	}
	return results
}

func indent(s string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = "  " + lines[i]
	}
	return strings.Join(lines, "\n")
}
