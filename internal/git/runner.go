package git

import "context"

type gitRunner struct{}

func (r gitRunner) AllBranches(ctx context.Context) ([]string, error) {
	return run(ctx, "branch", "--format=%(objectname:short) %(refname:short)")
}
