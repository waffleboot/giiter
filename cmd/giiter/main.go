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

	"github.com/urfave/cli/v2"
	"github.com/waffleboot/giiter/internal/config"
	"github.com/waffleboot/giiter/internal/git"
)

var (
	FlagRepo = &cli.StringFlag{
		Name:    "repo",
		Usage:   "path to git repository",
		Aliases: []string{"r"},
	}
	FlagBase = &cli.StringFlag{
		Name:     "base",
		Aliases:  []string{"b"},
		Usage:    "base branch",
		Required: true,
	}
	FlagFeat = &cli.StringFlag{
		Name:     "feature",
		Aliases:  []string{"f"},
		Usage:    "feature branch",
		Required: true,
	}
	FlagVerbose = &cli.BoolFlag{
		Name:    "verbose",
		Aliases: []string{"v"},
		Usage:   "verbose",
	}
	FlagDebug = &cli.BoolFlag{
		Name:    "debug",
		Aliases: []string{"d"},
		Usage:   "debug",
	}
	FlagPrefix = &cli.StringFlag{
		Name:    "prefix",
		Aliases: []string{"p"},
		Usage:   "merge review prefix",
	}
	FlagRefreshOnSubject = &cli.BoolFlag{
		Name:  "refresh-on-subj",
		Usage: "refresh using by subject",
	}
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app := &cli.App{
		Name: "giiter",
		Flags: []cli.Flag{
			FlagRepo,
			FlagDebug,
			FlagVerbose,
		},
		Before: func(*cli.Context) error {
			if err := config.LoadConfig(); err != nil {
				return err
			}
			return nil
		},
		After: func(*cli.Context) error {
			config.Close()
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "list feature commits",
				Aliases: []string{"l"},
				Action:  gitListFeatureCommits,
				Flags: []cli.Flag{
					FlagBase,
					FlagFeat,
					FlagRefreshOnSubject,
				},
				Before: func(ctx *cli.Context) error {
					config.Global.RefreshOnSubject = ctx.Bool("refresh-on-subj")
					return nil
				},
			},
			{
				Name:    "make",
				Usage:   "make review branches",
				Aliases: []string{"m"},
				Action:  gitMakeReviewBranches,
				Flags: []cli.Flag{
					FlagBase,
					FlagFeat,
					FlagPrefix,
				},
			},
			{
				Name:    "delete",
				Usage:   "delete review branches",
				Aliases: []string{"d"},
				Action:  gitDeleteReviewBranches,
				Flags: []cli.Flag{
					FlagFeat,
				},
			},
			{
				Name:    "assign",
				Usage:   "reassign commit to review branch",
				Action:  gitAssign,
				Aliases: []string{"a"},
				Flags: []cli.Flag{
					FlagBase,
					FlagFeat,
				},
			},
			{
				Name:    "branches",
				Usage:   "show all branches",
				Aliases: []string{"b"},
				Action:  gitShowAllBranches,
			},
		},
	}

	return app.RunContext(ctx, os.Args)
}

func gitListFeatureCommits(ctx *cli.Context) error {
	g := git.NewGit(ctx)

	base := ctx.String("base")
	feat := ctx.String("feature")

	records, err := g.Refresh(base, feat)
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

func gitMakeReviewBranches(ctx *cli.Context) error {
	g := git.NewGit(ctx)

	base := ctx.String("base")
	feat := ctx.String("feature")

	records, err := g.Refresh(base, feat)
	if err != nil {
		return err
	}

	for i := range records {
		if records[i].ReviewSHA != "" {
			continue
		}

		branch := fmt.Sprintf("review/%s/%d", feat, records[i].ID)

		targetBranch := base
		if i > 0 {
			targetBranch = records[i-1].ReviewBranch
		}

		title := "Draft: "
		if prefix := ctx.String("prefix"); prefix != "" {
			title += prefix + ": "
		}
		title += records[i].FeatureMsg.Subject

		if err := g.CreateBranch(
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

	return gitListFeatureCommits(ctx)
}

func gitDeleteReviewBranches(ctx *cli.Context) error {
	g := git.NewGit(ctx)

	branches, err := g.Branches()
	if err != nil {
		return err
	}

	prefix := fmt.Sprintf("review/%s/", ctx.String("feature"))

	for _, branch := range branches {
		if !strings.HasPrefix(branch.Name, prefix) {
			continue
		}
		if err := g.DeleteBranch(branch.Name); err != nil {
			return err
		}
	}

	return nil
}

func gitShowAllBranches(ctx *cli.Context) error {
	g := git.NewGit(ctx)

	branches, err := g.Branches()
	if err != nil {
		return err
	}

	for i := range branches {
		fmt.Printf("%s\n", branches[i].Name)
	}

	return nil
}

func gitAssign(ctx *cli.Context) error {
	if ctx.NArg() < 2 {
		return errors.New("need branch and commit")
	}

	sha, branch := ctx.Args().First(), ctx.Args().Get(1)

	branchIndex, err := strconv.Atoi(branch)
	if err != nil {
		return err
	}

	shaIndex, err := strconv.Atoi(sha)
	if err != nil {
		return err
	}

	g := git.NewGit(ctx)

	base := ctx.String("base")
	feat := ctx.String("feature")

	records, err := g.State(base, feat)
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

	if err := g.SwitchBranch(branch, sha); err != nil {
		return err
	}

	return gitListFeatureCommits(ctx)
}
