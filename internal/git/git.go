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
)

func AllBranches(ctx context.Context) ([]Branch, error) {
	output, err := run(ctx, "branch", "--format=%(objectname:short) %(refname:short)")
	if err != nil {
		return nil, errors.WithMessage(err, "get all branches")
	}

	branches := make([]Branch, 0, len(output))
	for _, line := range output {
		branch := Branch{
			CommitSHA:  line[:7],
			BranchName: line[8:],
		}
		branches = append(branches, branch)
	}

	return branches, nil
}

func DeleteBranch(ctx context.Context, name string) error {
	if name == "master" {
		panic(name)
	}
	_, err := run(ctx, "branch", "-D", name)
	if err != nil {
		return err
	}
	_, err = run(ctx, "push", "origin", "--delete", name)
	return err
}

type CreateBranchRequest struct {
	SHA         string
	Branch      string
	Target      string
	Title       string
	Description string
}

func CreateBranch(ctx context.Context, req CreateBranchRequest) error {
	if req.Branch == "master" {
		panic(req.Branch)
	}
	_, err := run(ctx, "branch", req.Branch, req.SHA)
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

	_, err = run(ctx, args...)
	return err
}

func Commits(ctx context.Context, baseBranch string) ([]string, error) {
	branches, err := AllBranches(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "get commits")
	}

	var baseFound, featFound bool
	for i := range branches {
		branch := branches[i].BranchName
		baseFound = baseFound || (branch == baseBranch)
		featFound = featFound || (branch == app.Config.FeatureBranch)
	}

	if !baseFound {
		return nil, errors.Errorf("branch '%s' not found", baseBranch)
	}

	if !featFound {
		return nil, errors.Errorf("branch '%s' not found", app.Config.FeatureBranch)
	}

	fromTo := fmt.Sprintf("%s..%s", baseBranch, app.Config.FeatureBranch)

	commits, err := run(ctx, "log", `--pretty=format:%h`, fromTo)
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

func SwitchBranch(ctx context.Context, branch, commit string) error {
	if branch == "master" {
		panic(branch)
	}
	_, err := run(ctx, "branch", "-f", branch, commit)
	if err != nil {
		return err
	}
	_, err = run(ctx, "push", "origin", "--force", branch+":"+branch)
	return err
}

func FindCommit(ctx context.Context, sha string) (*Commit, error) {
	output, err := run(ctx, "log", "--pretty=format:%s%n%b", sha, "-1")
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

func run(ctx context.Context, args ...string) ([]string, error) {
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

	if app.Config.Log != nil {
		fmt.Fprint(app.Config.Log, "git ")
		for i := range args {
			fmt.Fprint(app.Config.Log, args[i])
			fmt.Fprint(app.Config.Log, " ")
		}
		fmt.Fprintln(app.Config.Log)
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
