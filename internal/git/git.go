package git

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

type git struct {
	repo    string
	verbose bool
}

type Opts func(*git)

func Verbose() Opts {
	return func(g *git) {
		g.verbose = true
	}
}

func NewGit(repo string, opts ...Opts) *git {
	g := &git{repo: repo}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

func (g *git) run(ctx context.Context, args ...string) ([]string, error) {
	if g.verbose {
		fmt.Print("git")
		for i := range args {
			fmt.Printf(" %s", args[i])
		}
		fmt.Println()
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = g.repo

	stdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	s := bufio.NewScanner(bytes.NewReader(stdout))

	var out []string
	for s.Scan() {
		out = append(out, s.Text())
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (g *git) Branches(ctx context.Context) ([]string, error) {
	return g.run(ctx, "branch", "--format=%(refname:short)")
}

func (g *git) DeleteBranch(ctx context.Context, name string) error {
	_, err := g.run(ctx, "branch", "-d", name)
	return err
}

func (g *git) CreateBranch(ctx context.Context, name, base string) error {
	_, err := g.run(ctx, "branch", name, base)
	return err
}

func (g *git) Commits(ctx context.Context, base, feat string) ([]string, error) {
	return g.run(ctx, "log", `--pretty=format:%h`, fmt.Sprintf("%s..%s", base, feat))
}
