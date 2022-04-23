package git

import (
	"context"
	"errors"
	"fmt"

	"github.com/waffleboot/giiter/internal/app"
)

func getCurrentBranch(ctx context.Context) (string, error) {
	output, err := run(ctx, "branch", "--show-current")
	if err != nil {
		return "", err
	}

	return output[0], nil
}

func isProtectedBranch(branchName string) bool {
	return branchName == "main" || branchName == "master"
}

func FindFeatureBranch(ctx context.Context, featureBranch string) (string, error) {
	if featureBranch == "master" || featureBranch == "main" {
		return "", errors.New("master/main")
	}

	if featureBranch != "" {
		return featureBranch, nil
	}

	output, err := run(ctx, "branch", "--show-current")
	if err != nil {
		return "", err
	}

	return output[0], nil
}

func findBaseBranch(baseBranch, featureBranch string) string {
	for i := range app.Config.Persistent.FeatureBranches {
		item := app.Config.Persistent.FeatureBranches[i]
		if item.BranchName == featureBranch {
			if baseBranch != "" {
				app.Config.Persistent.FeatureBranches[i].BaseBranch = baseBranch
				return baseBranch
			}
			return item.BaseBranch
		}
	}
	if baseBranch != "" {
		app.Config.Persistent.FeatureBranches = append(app.Config.Persistent.FeatureBranches, app.FeatureBranch{
			BaseBranch: baseBranch,
			BranchName: featureBranch,
		})
	}
	return baseBranch
}

func FindBaseAndFeatureBranches(ctx context.Context, baseBranch, featureBranch string) (string, string, error) {
	featureBranch, err := FindFeatureBranch(ctx, featureBranch)
	if err != nil {
		return "", "", err
	}

	baseBranch = findBaseBranch(baseBranch, featureBranch)
	if baseBranch == "" {
		return "", "", errors.New("base branch is required")
	}

	return baseBranch, featureBranch, nil
}
