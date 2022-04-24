package main

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/waffleboot/giiter/internal/git"
)

var cmdDiff = &cobra.Command{
	Use:     "diff",
	Short:   "diff commit",
	Aliases: []string{"d"},
	// PersistentPreRunE не нужен, см. main
	RunE: showDiff,
}

func showDiff(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("point commit number")
	}

	shaPos := args[0]

	shaIndex, err := strconv.Atoi(shaPos)
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

	switch {
	case shaIndex < 0:
		return errors.New("commit position is negative")
	case shaIndex > len(records):
		return errors.New("commit position is greater then count of records")
	}

	record := records[shaIndex-1]

	lines, err := git.Diff(cmd.Context(), record.CommitSHA())
	if err != nil {
		return err
	}

	fmt.Println(record.CommitSHA(), record.CommitMessage().Subject)

	for _, line := range lines {
		fmt.Println(line)
	}

	return nil
}
