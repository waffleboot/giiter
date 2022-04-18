package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/waffleboot/giiter/internal/config"
	"github.com/waffleboot/giiter/internal/git"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

var (
	base             string
	feature          string
	refreshOnSubject bool
	prefix           string
	repo             string
	push             bool
	debug            bool
	verbose          bool
	g                *git.Git
)

var rootCmd = &cobra.Command{
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := config.LoadConfig(); err != nil {
			return err
		}
		g = &git.Git{
			Push:    push,
			Repo:    repo,
			Debug:   debug,
			Verbose: verbose,
		}
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		config.Close()
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "list feature commits",
	Aliases: []string{"l"},
	RunE:    gitListFeatureCommits,
}

var makeCmd = &cobra.Command{
	Use:     "make",
	Short:   "make review branches",
	Aliases: []string{"m"},
	RunE:    gitMakeReviewBranches,
}

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "delete review branches",
	Aliases: []string{"d"},
	RunE:    gitDeleteReviewBranches,
}

var assignCmd = &cobra.Command{
	Use:     "assign",
	Short:   "reassign commit to review branch",
	Aliases: []string{"a"},
	RunE:    gitAssign,
}

var branchesCmd = &cobra.Command{
	Use:     "branches",
	Short:   "show all branches",
	Aliases: []string{"b"},
	RunE:    gitShowAllBranches,
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rootCmd.PersistentFlags().StringVar(&repo, "repo", "", "path to git repository")
	rootCmd.PersistentFlags().BoolVar(&push, "push", false, "git push")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose")

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(makeCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(assignCmd)
	rootCmd.AddCommand(branchesCmd)

	listCmd.Flags().StringVarP(&base, "base", "b", "", "base branch")
	listCmd.MarkFlagRequired("base")

	listCmd.Flags().StringVarP(&feature, "feature", "f", "", "feature branch")
	listCmd.MarkFlagRequired("feature")

	listCmd.Flags().BoolVar(&refreshOnSubject, "refresh-on-subj", false, "refresh using by subject")

	makeCmd.Flags().StringVarP(&base, "base", "b", "", "base branch")
	makeCmd.MarkFlagRequired("base")

	makeCmd.Flags().StringVarP(&feature, "feature", "f", "", "feature branch")
	makeCmd.MarkFlagRequired("feature")

	makeCmd.Flags().StringVar(&prefix, "prefix", "", "merge review prefix")

	deleteCmd.Flags().StringVarP(&feature, "feature", "f", "", "feature branch")
	deleteCmd.MarkFlagRequired("feature")

	assignCmd.Flags().StringVarP(&base, "base", "b", "", "base branch")
	assignCmd.MarkFlagRequired("base")

	assignCmd.Flags().StringVarP(&feature, "feature", "f", "", "feature branch")
	assignCmd.MarkFlagRequired("feature")

	return rootCmd.ExecuteContext(ctx)
}

func gitListFeatureCommits(cmd *cobra.Command, args []string) error {
	records, err := g.Refresh(cmd.Context(), base, feature)
	if err != nil {
		return err
	}

	for i := range records {
		record := records[i]
		if record.IsNewCommit() {
			fmt.Printf("%d) + %s %s\n", i+1, record.FeatureSHA, record.FeatureMsg.Subject)
		} else if record.IsOldCommit() {
			fmt.Printf("%d) - %s [%s] %s\n", i+1, record.ReviewSHA, record.ReviewBranch, record.ReviewMsg.Subject)
		} else {
			fmt.Printf("%d) . %s [%s] %s\n", i+1, record.FeatureSHA, record.ReviewBranch, record.FeatureMsg.Subject)
		}
	}

	return nil
}

func gitMakeReviewBranches(cmd *cobra.Command, args []string) error {
	records, err := g.Refresh(cmd.Context(), base, feature)
	if err != nil {
		return err
	}

	for i := range records {
		if records[i].ReviewSHA != "" {
			continue
		}

		branch := fmt.Sprintf("review/%s/%d", feature, records[i].ID)

		targetBranch := base
		if i > 0 {
			targetBranch = records[i-1].ReviewBranch
		}

		title := "Draft: "
		if prefix != "" {
			title += prefix + ": "
		}
		title += records[i].FeatureMsg.Subject

		if err := g.CreateBranch(
			cmd.Context(),
			git.CreateBranchRequest{
				SHA:         records[i].FeatureSHA,
				Branch:      branch,
				Target:      targetBranch,
				Title:       title,
				Description: records[i].FeatureMsg.Description,
			}); err != nil {
			return err
		}

		records[i].ReviewBranch = branch

	}

	return gitListFeatureCommits(cmd, args)
}

func gitDeleteReviewBranches(cmd *cobra.Command, args []string) error {
	branches, err := g.Branches(cmd.Context())
	if err != nil {
		return err
	}

	prefix := fmt.Sprintf("review/%s/", feature)

	for _, branch := range branches {
		if !strings.HasPrefix(branch.Name, prefix) {
			continue
		}
		if err := g.DeleteBranch(cmd.Context(), branch.Name); err != nil {
			return err
		}
	}

	return nil
}

func gitShowAllBranches(cmd *cobra.Command, args []string) error {
	branches, err := g.Branches(cmd.Context())
	if err != nil {
		return err
	}

	for i := range branches {
		fmt.Printf("%s\n", branches[i].Name)
	}

	return nil
}

func gitAssign(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return errors.New("need branch and commit")
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

	records, err := g.State(cmd.Context(), base, feature)
	if err != nil {
		return err
	}

	branch = records[branchIndex-1].ReviewBranch
	sha = records[shaIndex-1].FeatureSHA

	reviewIndex := -1
	commitIndex := -1

	for i := range records {
		if records[i].ReviewBranch == branch {
			reviewIndex = i
		}
		if records[i].FeatureSHA == sha {
			commitIndex = i
		}
	}

	if reviewIndex < 0 || commitIndex < 0 || reviewIndex == commitIndex {
		return nil
	}

	if err := g.SwitchBranch(cmd.Context(), branch, sha); err != nil {
		return err
	}

	return gitListFeatureCommits(cmd, args)
}
