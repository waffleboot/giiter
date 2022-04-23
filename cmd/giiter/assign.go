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

	branch = records[branchIndex-1].ReviewBranch.BranchName
	sha = records[shaIndex-1].FeatureSHA

	reviewIndex := -1
	commitIndex := -1

	for i := range records {
		if records[i].ReviewBranch.BranchName == branch {
			reviewIndex = i
		}

		if records[i].FeatureSHA == sha {
			commitIndex = i
		}
	}

	if reviewIndex < 0 || commitIndex < 0 || reviewIndex == commitIndex {
		return nil
	}

	if err := git.SwitchBranch(cmd.Context(), branch, sha); err != nil {
		return err
	}

	return listFeatureCommits(cmd, args)
}
