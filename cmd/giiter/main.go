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
		Commands: []*cli.Command{
			{
				Name:    "git",
				Aliases: []string{"g"},
				Subcommands: []*cli.Command{
					{
						Name:    "list",
						Aliases: []string{"l"},
						Action:  gitList,
						Flags: []cli.Flag{
							FlagBase,
							FlagFeat,
						},
					},
					{
						Name:    "make",
						Aliases: []string{"m"},
						Action:  gitMake,
						Flags: []cli.Flag{
							FlagBase,
							FlagFeat,
						},
					},
					{
						Name:    "clear",
						Aliases: []string{"c"},
						Action:  gitClear,
						Flags: []cli.Flag{
							FlagFeat,
						},
					},
					{
						Name:    "branches",
						Aliases: []string{"b"},
						Action:  gitBranches,
					},
					{
						Name:   "check",
						Action: gitCheck,
						Flags: []cli.Flag{
							FlagBase,
							FlagFeat,
						},
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

func gitList(ctx *cli.Context) error {
	g := git.NewGit(ctx.String("repo"), git.Verbose())
	commits, err := g.Commits(ctx.Context, ctx.String("base"), ctx.String("feature"))
	if err != nil {
		return err
	}
	for i := range commits {
		fmt.Println(commits[i])
	}
	return nil
}

// type maker struct {
// 	repo string
// 	base string
// 	feat string
// 	sha  string
// }

// func (m *maker) make(ctx *cli.Context) error {
// 	return nil
// }

func gitMake(ctx *cli.Context) error {
	g := git.NewGit(ctx.String("repo"), git.Verbose())

	base := ctx.String("base")
	feat := ctx.String("feature")

	commits, err := g.Commits(ctx.Context, base, feat)
	if err != nil {
		return err
	}

	prefix := reviewBranchPrefix(feat)

	for i := range commits {
		branch := prefix + strconv.Itoa(i+1)
		if err := g.CreateBranch(ctx.Context, branch, commits[i]); err != nil {
			return err
		}
		// if _, err := g.run(ctx.Context, "branch", branch, sha); err != nil {
		// 	return err
		// }
		// if _, err := g.run(ctx.Context, "reset", "--hard", branch, sha+"~1", sha); err != nil {
		// 	return err
		// }
		// if _, err := g.run(ctx.Context, "rebase", "--onto", base, sha, branch); err != nil {
		// 	return err
		// }
	}

	// 		// git branch feat-sha base
	// 		// git cherry pick sha

	return nil
}

func gitClear(ctx *cli.Context) error {
	g := git.NewGit(ctx.String("repo"), git.Verbose())

	feat := ctx.String("feature")
	if feat == "" {
		return errors.New("feature is required")
	}

	branches, err := g.Branches(ctx.Context)
	if err != nil {
		return err
	}

	prefix := reviewBranchPrefix(feat)

	for _, branch := range branches {
		name := branch.Name
		if strings.HasPrefix(name, prefix) {
			if err := g.DeleteBranch(ctx.Context, name); err != nil {
				return err
			}
		}
	}

	return nil
}

func gitBranches(ctx *cli.Context) error {
	g := git.NewGit(ctx.String("repo"), git.Verbose())

	out, err := g.Branches(ctx.Context)
	if err != nil {
		return err
	}

	for i := range out {
		fmt.Println(out[i])
	}

	return nil
}

func gitCheck(ctx *cli.Context) error {
	g := git.NewGit(ctx.String("repo"), git.Verbose())

	base := ctx.String("base")
	feat := ctx.String("feature")

	hashToCommit := make(map[string]string)

	commits, err := g.Commits(ctx.Context, base, feat)
	if err != nil {
		return err
	}

	for _, commit := range commits {
		hash, err := g.DiffHash(ctx.Context, commit)
		if err != nil {
			return err
		}
		hashToCommit[fmt.Sprintf("%x", hash)] = commit
	}

	branches, err := g.Branches(ctx.Context)
	if err != nil {
		return err
	}

	prefix := reviewBranchPrefix(feat)

	hashToBranch := make(map[string]string)

	for _, branch := range branches {
		if strings.HasPrefix(branch.Name, prefix) {

			hash, err := g.DiffHash(ctx.Context, branch.Commit)
			if err != nil {
				return err
			}

			hashToBranch[fmt.Sprintf("%x", hash)] = branch.Name
		}
	}

	for k, v := range hashToCommit {
		fmt.Println(k, v)
	}

	fmt.Println("----")

	for k, v := range hashToBranch {
		fmt.Println(k, v)
	}

	var fixBranches []string

	for hash, branch := range hashToBranch {
		if commit, ok := hashToCommit[hash]; ok {
			g.SwitchBranch(ctx.Context, branch, commit)
		} else {
			fixBranches = append(fixBranches, branch)
		}
	}

	var fixCommits []string
	for hash, commit := range hashToCommit {
		if _, ok := hashToBranch[hash]; !ok {
			fixCommits = append(fixCommits, commit)
		}
	}

	if len(fixBranches) > 0 {
		fmt.Println("FIX BRANCHES")
		for _, branch := range fixBranches {
			fmt.Printf("git branch -f %s \n", branch)
		}
	}

	if len(fixCommits) > 0 {
		fmt.Println("FIX COMMITS")
		for _, commit := range fixCommits {
			fmt.Println(commit)
		}
	}

	return nil
}

func reviewBranchPrefix(feature string) string {
	return fmt.Sprintf("review/%s/", feature)
}
