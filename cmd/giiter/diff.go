package main

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/waffleboot/giiter/internal/git"
)

type diffCommand struct {
	config *git.Config
}

func makeDiffCommand(config *git.Config) *cobra.Command {
	c := diffCommand{
		config: config,
	}

	return &cobra.Command{
		Use:     "diff",
		Short:   "diff commit",
		Aliases: []string{"d"},
		Args:    cobra.MinimumNArgs(1),
		// PersistentPreRunE не нужен, см. main
		RunE: c.run,
	}
}

func (c *diffCommand) run(cmd *cobra.Command, args []string) error {
	baseBranch, featureBranch, err := c.config.Branches()
	if err != nil {
		return err
	}

	shaPos := args[0]

	shaIndex, err := strconv.Atoi(shaPos)
	if err != nil {
		return err
	}

	records, err := git.State(cmd.Context(), baseBranch, featureBranch)
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
