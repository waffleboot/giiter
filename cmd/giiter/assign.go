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

	shaPos, branchPos := args[0], args[1]

	branchIndex, err := strconv.Atoi(branchPos)
	if err != nil {
		return err
	}

	shaIndex, err := strconv.Atoi(shaPos)
	if err != nil {
		return err
	}

	if branchIndex == shaIndex {
		return errors.New("you point the same record")
	}

	records, err := git.State(
		cmd.Context(),
		_baseBranch,
		_featureBranch)
	if err != nil {
		return err
	}

	branchName := records[branchIndex-1].ReviewBranch.BranchName
	featureSHA := records[shaIndex-1].FeatureSHA

	reviewIndex := -1
	commitIndex := -1

	for i := range records {
		if records[i].ReviewBranch.BranchName == branchName {
			reviewIndex = i
		}

		if records[i].FeatureSHA == featureSHA {
			commitIndex = i
		}
	}

	// TODO зачем это сделано?

	if reviewIndex < 0 || commitIndex < 0 || reviewIndex == commitIndex {
		return nil
	}

	if err := git.SwitchBranch(cmd.Context(), branchName, featureSHA); err != nil {
		return err
	}

	return listFeatureCommits(cmd, args)
}
