package git

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
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

type Branch struct {
	Name   string
	Commit string
}

func (g *git) Branches(ctx context.Context) ([]Branch, error) {
	out, err := g.run(ctx, "branch", "--format=%(objectname:short) %(refname:short)")
	if err != nil {
		return nil, err
	}

	result := make([]Branch, 0, len(out))

	for _, line := range out {
		obj := Branch{
			Commit: line[:7],
			Name:   line[8:],
		}
		result = append(result, obj)
	}

	return result, nil
}

func (g *git) DeleteBranch(ctx context.Context, name string) error {
	_, err := g.run(ctx, "branch", "-D", name)
	return err
}

func (g *git) CreateBranch(ctx context.Context, name, base string) error {
	_, err := g.run(ctx, "branch", name, base)
	return err
}

func (g *git) Commits(ctx context.Context, base, feat string) ([]string, error) {
	commits, err := g.run(ctx, "log", `--pretty=format:%h`, fmt.Sprintf("%s..%s", base, feat))
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(commits)/2; i++ {
		r := len(commits) - i - 1
		commits[i], commits[r] = commits[r], commits[i]
	}
	return commits, err
}

func (g *git) DiffHash(ctx context.Context, commit string) ([]byte, error) {
	hash := sha256.New()

	diff, err := g.run(ctx, "diff", "--unified=0", commit+"~", commit)
	if err != nil {
		return nil, err
	}
	diff = diff[2:]

	for _, line := range diff {
		hash.Write([]byte(line))
	}

	sum := hash.Sum(nil)

	if g.verbose {
		fmt.Println("--- diff ---")
		for _, d := range diff {
			fmt.Println(d)
		}
		fmt.Printf("%x\n", sum)
		fmt.Println("--- diff ---")
	}

	return sum, nil
}

func (g *git) SwitchBranch(ctx context.Context, branch, commit string) error {
	_, err := g.run(ctx, "branch", "-f", branch, commit)
	return err
}
