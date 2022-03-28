package git

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/waffleboot/giiter/internal/config"
)

type git struct {
	repo    string
	debug   bool
	verbose bool
	context context.Context
}

type Opts func(*git)

func NewGit(ctx *cli.Context) *git {
	g := &git{
		repo:    ctx.String("repo"),
		debug:   ctx.Bool("debug"),
		verbose: ctx.Bool("verbose"),
		context: ctx.Context,
	}
	return g
}

func (g *git) Branches() ([]Branch, error) {
	output, err := g.run("branch", "--format=%(objectname:short) %(refname:short)")
	if err != nil {
		return nil, err
	}

	branches := make([]Branch, 0, len(output))
	for _, line := range output {
		branch := Branch{
			SHA:  line[:7],
			Name: line[8:],
		}
		branches = append(branches, branch)
	}

	return branches, nil
}

func (g *git) DeleteBranch(name string) error {
	if name == "master" {
		panic(name)
	}
	_, err := g.run("branch", "-D", name)
	if err != nil {
		return err
	}
	_, err = g.run("push", "origin", "--delete", name)
	return err
}

type CreateBranchRequest struct {
	SHA         string
	Branch      string
	Target      string
	Title       string
	Description string
}

func (g *git) CreateBranch(req CreateBranchRequest) error {
	if req.Branch == "master" {
		panic(req.Branch)
	}
	_, err := g.run("branch", req.Branch, req.SHA)
	if err != nil {
		return err
	}

	args := []string{
		"push",
		"-o", "merge_request.create",
		"-o", "merge_request.target=" + req.Target,
		"-o", "merge_request.title=" + req.Title,
		"-o", "merge_request.label=review",
	}

	if req.Description != "" {
		args = append(args, "-o", "merge_request.description="+req.Description)
	}

	args = append(args, "origin", req.Branch+":"+req.Branch)

	_, err = g.run(args...)
	return err
}

func (g *git) Commits(base, feat string) ([]string, error) {
	fromTo := fmt.Sprintf("%s..%s", base, feat)

	commits, err := g.run("log", `--pretty=format:%h`, fromTo)
	if err != nil {
		return nil, err
	}

	// reverse order
	for i := 0; i < len(commits)/2; i++ {
		r := len(commits) - i - 1
		commits[i], commits[r] = commits[r], commits[i]
	}

	return commits, err
}

func (g *git) SwitchBranch(branch, commit string) error {
	if branch == "master" {
		panic(branch)
	}
	_, err := g.run("branch", "-f", branch, commit)
	if err != nil {
		return err
	}
	_, err = g.run("push", "origin", "--force", branch+":"+branch)
	return err
}

func (g *git) FindCommit(sha string) (*Commit, error) {
	output, err := g.run("log", "--pretty=format:%s%n%b", sha, "-1")
	if err != nil {
		return nil, err
	}

	var body string
	if len(output) > 1 {
		body = strings.Join(output[1:], "\n")
	}

	commit := &Commit{
		SHA: sha,
		Message: Message{
			Subject:     output[0],
			Description: body,
		},
	}

	return commit, nil
}

func (g *git) run(args ...string) ([]string, error) {
	if g.verbose {
		fmt.Print("git")
		for i := range args {
			fmt.Printf(" %s", args[i])
		}
		fmt.Println()
	}

	if config.Config.Log != nil {
		fmt.Fprint(config.Config.Log, "git ")
		for i := range args {
			fmt.Fprint(config.Config.Log, args[i])
			fmt.Fprint(config.Config.Log, " ")
		}
		fmt.Fprintln(config.Config.Log)
	}

	cmd := exec.CommandContext(g.context, "git", args...)

	cmd.Dir = g.repo

	stdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var output []string

	scanner := bufio.NewScanner(bytes.NewReader(stdout))

	for scanner.Scan() {
		output = append(output, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return output, nil
}
