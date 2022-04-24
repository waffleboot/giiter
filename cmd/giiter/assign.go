package main

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/waffleboot/giiter/internal/git"
)

var assignCmd = &cobra.Command{
	Use:     "assign",
	Short:   "reassign commit to review branch",
	Aliases: []string{"a"},
	// PersistentPreRunE не нужен, см. main
	RunE: assign,
}

func assign(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return errors.New("need new commit and old review branch position numbers")
	}

	sha, branch := args[0], args[1]

	branchIndex, err := strconv.Atoi(branch)
	if err != nil {
		return err
	}

	shaIndex, err := strconv.Atoi(sha)
	if err != nil {
		return err
	}

	records, err := git.State(
		cmd.Context(),
		_baseBranch,
		_featureBranch)
	if err != nil {
		return err
	}

	if shaIndex == branchIndex {
		return nil
	}

	sha = records[shaIndex-1].FeatureSHA
	branch = records[branchIndex-1].MainReviewBranch()

	if err := git.SwitchBranch(cmd.Context(), branch, sha); err != nil {
		return err
	}

	return listFeatureCommits(cmd, args)
}
