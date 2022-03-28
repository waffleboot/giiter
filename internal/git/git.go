package git

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/urfave/cli/v2"
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

func (g *git) CreateBranch(name, sha, targetBranch, title string) error {
	if name == "master" {
		panic(name)
	}
	_, err := g.run("branch", name, sha)
	if err != nil {
		return err
	}
	_, err = g.run("push",
		"-o", "merge_request.create",
		"-o", "merge_request.target="+targetBranch,
		"-o", "merge_request.title=DRAFT: "+title,
		"-o", "merge_request.label=review",
		"origin", name+":"+name)
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
	output, err := g.run("log", "--pretty=format:%s", sha, "-1")
	if err != nil {
		return nil, err
	}

	commit := &Commit{
		SHA:     sha,
		Subject: output[0],
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
