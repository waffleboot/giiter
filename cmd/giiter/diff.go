package main

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/waffleboot/giiter/internal/git"
)

type diffCommand struct {
	branches *branches
}

func makeDiffCommand(branches *branches) *cobra.Command {
	c := diffCommand{
		branches: branches,
	}

	return &cobra.Command{
		Use:     "diff",
		Short:   "diff commit",
		Aliases: []string{"d"},
		// PersistentPreRunE не нужен, см. main
		RunE: c.run,
	}
}

func (c *diffCommand) run(cmd *cobra.Command, args []string) error {
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
		c.branches.baseBranch,
		c.branches.featureBranch)
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

	fmt.Println(Yellow + "commit " + record.CommitSHA() + " message " + record.CommitMessage().Subject + Reset)

	if err := git.Diff(cmd.Context(), record.CommitSHA(), args[1:]...); err != nil {
		return err
	}

	return nil
}
