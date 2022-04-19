package git

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/waffleboot/giiter/internal/app"
	"github.com/waffleboot/giiter/internal/config"
)

type Git struct{}

func (g *Git) Branches(ctx context.Context) ([]Branch, error) {
	output, err := g.run(ctx, "branch", "--format=%(objectname:short) %(refname:short)")
	if err != nil {
		return nil, errors.WithMessage(err, "get all branches")
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

func (g *Git) DeleteBranch(ctx context.Context, name string) error {
	if name == "master" {
		panic(name)
	}
	_, err := g.run(ctx, "branch", "-D", name)
	if err != nil {
		return err
	}
	_, err = g.run(ctx, "push", "origin", "--delete", name)
	return err
}

type CreateBranchRequest struct {
	SHA         string
	Branch      string
	Target      string
	Title       string
	Description string
}

func (g *Git) CreateBranch(ctx context.Context, req CreateBranchRequest) error {
	if req.Branch == "master" {
		panic(req.Branch)
	}
	_, err := g.run(ctx, "branch", req.Branch, req.SHA)
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

	_, err = g.run(ctx, args...)
	return err
}

func (g *Git) Commits(ctx context.Context) ([]string, error) {
	branches, err := g.Branches(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "get commits")
	}

	var baseFound, featFound bool
	for i := range branches {
		branch := branches[i].Name
		baseFound = baseFound || (branch == app.Config.BaseBranch)
		featFound = featFound || (branch == app.Config.FeatureBranch)
	}

	if !baseFound {
		return nil, errors.Errorf("branch '%s' not found", app.Config.BaseBranch)
	}

	if !featFound {
		return nil, errors.Errorf("branch '%s' not found", app.Config.FeatureBranch)
	}

	fromTo := fmt.Sprintf("%s..%s", app.Config.BaseBranch, app.Config.FeatureBranch)

	commits, err := g.run(ctx, "log", `--pretty=format:%h`, fromTo)
	if err != nil {
		return nil, errors.WithMessage(err, "get commits by log")
	}

	// reverse order
	for i := 0; i < len(commits)/2; i++ {
		r := len(commits) - i - 1
		commits[i], commits[r] = commits[r], commits[i]
	}

	return commits, err
}

func (g *Git) SwitchBranch(ctx context.Context, branch, commit string) error {
	if branch == "master" {
		panic(branch)
	}
	_, err := g.run(ctx, "branch", "-f", branch, commit)
	if err != nil {
		return err
	}
	_, err = g.run(ctx, "push", "origin", "--force", branch+":"+branch)
	return err
}

func (g *Git) FindCommit(ctx context.Context, sha string) (*Commit, error) {
	output, err := g.run(ctx, "log", "--pretty=format:%s%n%b", sha, "-1")
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

func (g *Git) run(ctx context.Context, args ...string) ([]string, error) {
	if !app.Config.Push && args[0] == "push" {
		return nil, nil
	}
	if app.Config.Verbose {
		fmt.Print("git")
		for i := range args {
			fmt.Printf(" %s", args[i])
		}
		fmt.Println()
	}

	if config.Global.Log != nil {
		fmt.Fprint(config.Global.Log, "git ")
		for i := range args {
			fmt.Fprint(config.Global.Log, args[i])
			fmt.Fprint(config.Global.Log, " ")
		}
		fmt.Fprintln(config.Global.Log)
	}

	cmd := exec.CommandContext(ctx, "git", args...)

	cmd.Dir = app.Config.Repo

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
