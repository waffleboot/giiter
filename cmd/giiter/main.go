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
			FlagVerbose,
			FlagDebug,
		},
		Commands: []*cli.Command{
			{
				Name:    "git",
				Aliases: []string{"g"},
				Subcommands: []*cli.Command{
					{
						Name:    "list",
						Usage:   "list feature commits",
						Aliases: []string{"l"},
						Action:  gitListFeatureCommits,
						Flags: []cli.Flag{
							FlagBase,
							FlagFeat,
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
				Flags: []cli.Flag{
					FlagRepo,
				},
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
		if record.ReviewSHA == "" {
			fmt.Printf("%d) + %s %s\n", i+1, record.FeatureSHA, record.FeatureSubj)
		} else if record.FeatureSHA == "" {
			fmt.Printf("%d) - %s [%s] %s\n", i+1, record.ReviewSHA, record.ReviewBranch, record.ReviewSubj)
		} else {
			fmt.Printf("%d)   %s [%s] %s\n", i+1, record.FeatureSHA, record.ReviewBranch, record.FeatureSubj)
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
		branch := fmt.Sprintf("review/%s/%d", feat, i+1)
		if err := g.CreateBranch(branch, records[i].FeatureSHA); err != nil {
			return err
		}
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

// func gitCheckFeatureStack(ctx *cli.Context) error {

// 	if len(fixBranches) > 0 {
// 		fmt.Println("FIX BRANCHES")
// 		for _, branch := range fixBranches {
// 			commit, err := g.FindCommit(branch.Name)
// 			if err != nil {
// 				return err
// 			}
// 			fmt.Printf("git branch -f %s Subj:%s\n", branch.Name, commit.Subject)
// 		}
// 	}

// 	if len(fixCommits) > 0 {
// 		fmt.Println("FIX COMMITS")
// 		for _, sha := range fixCommits {
// 			commit, err := g.FindCommit(sha)
// 			if err != nil {
// 				return err
// 			}
// 			fmt.Printf("%s %s\n", commit.SHA, commit.Subject)
// 		}
// 	}

// 	return nil
// }

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
