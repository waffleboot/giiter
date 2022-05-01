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

func DeleteBranch(ctx context.Context, branchName string) error {
	if isProtectedBranch(branchName) {
		return fmt.Errorf("%s is proteced branch, could not delete it", branchName)
	}

	_, err := run(ctx, "branch", "-D", branchName)
	if err != nil {
		return err
	}

	_, err = run(ctx, "push", "origin", "--delete", branchName)

	return err
}

func CreateBranch(ctx context.Context, branch Branch) error {
	if isProtectedBranch(branch.BranchName) {
		return fmt.Errorf("%s is protected branch, could not create it", branch.BranchName)
	}

	_, err := run(ctx, "branch", branch.BranchName, branch.CommitSHA)

	return err
}

type MergeRequest struct {
	Title        string
	SourceBranch string
	TargetBranch string
	Description  string
}

func CreateMergeRequest(ctx context.Context, req MergeRequest) error {
	if isProtectedBranch(req.SourceBranch) {
		return fmt.Errorf("%s is protected branch, merge requests disabled", req.SourceBranch)
	}

	args := []string{
		"push",
		"-o", "merge_request.create",
		"-o", "merge_request.target=" + req.TargetBranch,
		"-o", "merge_request.title=" + req.Title,
		"-o", "merge_request.label=review",
	}

	if req.Description != "" {
		args = append(args, "-o", "merge_request.description="+req.Description)
	}

	args = append(args, "origin", req.SourceBranch+":"+req.SourceBranch)

	_, err := run(ctx, args...)

	return err
}

func Commits(ctx context.Context, baseBranch, featureBranch string) ([]string, error) {
	branches, err := AllBranches(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "get commits")
	}

	var baseFound, featFound bool

	for i := range branches {
		branch := branches[i].BranchName
		baseFound = baseFound || (branch == baseBranch)
		featFound = featFound || (branch == featureBranch)
	}

	if !baseFound {
		return nil, errors.Errorf("branch '%s' not found", baseBranch)
	}

	if !featFound {
		return nil, errors.Errorf("branch '%s' not found", featureBranch)
	}

	interval := fmt.Sprintf("%s..%s", baseBranch, featureBranch)

	commits, err := run(ctx, "log", `--pretty=format:%h`, "--first-parent", interval)
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
	if isProtectedBranch(branch) {
		return fmt.Errorf("%s is protected branch, disable switch", branch)
	}

	_, err := run(ctx, "branch", "-f", branch, commit)
	if err != nil {
		return err
	}

	_, err = run(ctx, "push", "origin", "--force", branch+":"+branch)

	return err
}

func FindCommit(ctx context.Context, sha string) (*commit, error) {
	output, err := run(ctx, "log", "--pretty=format:%s%n%b", sha, "-1")
	if err != nil {
		return nil, err
	}

	var body string
	if len(output) > 1 {
		body = strings.Join(output[1:], "\n")
	}

	commit := &commit{
		SHA: sha,
		Message: Message{
			Subject:     output[0],
			Description: body,
		},
	}

	return commit, nil
}

func Rebase(ctx context.Context, baseBranch, featureBranch string) error {
	fmt.Printf("git rebase --onto %s %s %s\n", baseBranch, baseBranch, featureBranch)

	_, errRebase := run(ctx, "rebase", "--onto", baseBranch, baseBranch, featureBranch)
	if errRebase != nil {
		var errRun ErrRun
		if errors.As(errRebase, &errRun) {
			errRun.log()
		}

		_, errAbort := run(ctx, "rebase", "--abort")
		if errors.As(errAbort, &errRun) {
			errRun.log()
		}

		return errRebase
	}

	return nil
}

type ErrRun struct {
	stdOutput []string
	errOutput []string
	err       error
}

func (e ErrRun) Error() string {
	return e.err.Error()
}

func (e ErrRun) log() {
	for i := range e.stdOutput {
		fmt.Println(e.stdOutput[i])
	}

	for i := range e.errOutput {
		fmt.Println(e.errOutput[i])
	}
}

func run(ctx context.Context, args ...string) ([]string, error) {
	if !app.Config.EnableGitPush && args[0] == "push" {
		return nil, nil
	}

	if app.Config.Verbose {
		fmt.Print("git")

		for i := range args {
			fmt.Printf(" %s", args[i])
		}

		fmt.Println()
	}

	// if app.Config.Log != nil {
	// 	fmt.Fprint(app.Config.Log, "git ")
	// 	for i := range args {
	// 		fmt.Fprint(app.Config.Log, args[i])
	// 		fmt.Fprint(app.Config.Log, " ")
	// 	}
	// 	fmt.Fprintln(app.Config.Log)
	// }

	cmd := exec.CommandContext(ctx, "git", args...)

	// cmd.Dir = app.Config.Repo

	stdOut := new(bytes.Buffer)
	stdErr := new(bytes.Buffer)

	cmd.Stdout = stdOut
	cmd.Stderr = stdErr

	errRun := cmd.Run()

	stdOutLines, errParseStdOut := bytesBufferToSlice(stdOut)
	stdErrLines, errParseErrOut := bytesBufferToSlice(stdErr)

	if errParseStdOut != nil {
		fmt.Println(errParseStdOut)
	}

	if errParseErrOut != nil {
		fmt.Println(errParseErrOut)
	}

	if errRun != nil {
		return nil, ErrRun{
			stdOutput: stdOutLines,
			errOutput: stdErrLines,
			err:       errRun,
		}
	}

	return stdOutLines, nil
}

func bytesBufferToSlice(buf *bytes.Buffer) ([]string, error) {
	var output []string

	scanner := bufio.NewScanner(buf)

	for scanner.Scan() {
		output = append(output, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return output, nil
}
