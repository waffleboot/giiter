package git

import "context"

type runner struct{}

func (r runner) AllBranches(ctx context.Context) ([]string, error) {
	return run(ctx, "branch", "--format=%(objectname:short) %(refname:short)")
}

func (r runner) ChangedFiles(ctx context.Context, sha string) ([]string, error) {
	return run(ctx, "diff-tree", "-r", "--name-only", "-c", sha)
}
