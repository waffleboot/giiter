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
						Name:    "clear",
						Usage:   "delete review branches",
						Aliases: []string{"c"},
						Action:  gitDeleteReviewBranches,
						Flags: []cli.Flag{
							FlagFeat,
						},
					},
					{
						Name:    "branches",
						Usage:   "show all branches",
						Aliases: []string{"b"},
						Action:  gitShowAllBranches,
					},
					{
						Name:   "check",
						Usage:  "check and update feature branches",
						Action: gitCheckFeatureStack,
						Flags: []cli.Flag{
							FlagBase,
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

	commits, err := g.Commits(base, feat)
	if err != nil {
		return err
	}

	for i := range commits {
		commit, err := g.FindCommit(commits[i])
		if err != nil {
			return err
		}

		fmt.Printf("%s %s\n", commit.SHA, commit.Subject)
	}

	return nil
}

func gitMakeReviewBranches(ctx *cli.Context) error {
	_, fixCommits, err := check(ctx)
	if err != nil {
		return err
	}

	if len(fixCommits) == 0 {
		return nil
	}

	g := git.NewGit(ctx)

	branches, err := g.Branches()
	if err != nil {
		return err
	}

	feat := ctx.String("feature")

	prefix := reviewBranchPrefix(feat)

	featureBranches := make(map[string]struct{})

	for _, branch := range branches {
		name := branch.Name
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		featureBranches[branch.Name] = struct{}{}
	}

	var n int

	for i := range fixCommits {
		var branch string
		for {
			n++
			branch = prefix + strconv.Itoa(n)
			if _, exists := featureBranches[branch]; !exists {
				break
			}
		}
		if err := g.CreateBranch(branch, fixCommits[i]); err != nil {
			return err
		}
	}

	return nil
}

func gitDeleteReviewBranches(ctx *cli.Context) error {
	g := git.NewGit(ctx)

	feat := ctx.String("feature")

	branches, err := g.Branches()
	if err != nil {
		return err
	}

	prefix := reviewBranchPrefix(feat)

	for _, branch := range branches {
		name := branch.Name
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		if err := g.DeleteBranch(name); err != nil {
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

func check(ctx *cli.Context) ([]git.Branch, []string, error) {
	g := git.NewGit(ctx)

	base := ctx.String("base")
	feat := ctx.String("feature")

	commits, err := g.Commits(base, feat)
	if err != nil {
		return nil, nil, err
	}

	hashToCommit := make(map[string]string)
	for _, commit := range commits {
		hash, err := g.DiffHash(commit)
		if err != nil {
			return nil, nil, err
		}
		hashToCommit[hash] = commit
	}

	branches, err := g.Branches()
	if err != nil {
		return nil, nil, err
	}

	prefix := reviewBranchPrefix(feat)

	branchToSHA := make(map[string]string)
	hashToBranch := make(map[string]string)
	for _, branch := range branches {
		if strings.HasPrefix(branch.Name, prefix) {

			hash, err := g.DiffHash(branch.SHA)
			if err != nil {
				return nil, nil, err
			}

			hashToBranch[hash] = branch.Name
			branchToSHA[branch.Name] = branch.SHA
		}
	}

	// for k, v := range hashToCommit {
	// 	fmt.Println(k, v)
	// }

	// fmt.Println("----")

	// for k, v := range hashToBranch {
	// 	fmt.Println(k, v)
	// }

	var fixBranches []git.Branch
	for hash, branch := range hashToBranch {
		if commit, ok := hashToCommit[hash]; ok {
			g.SwitchBranch(branch, commit)
		} else {
			fixBranches = append(fixBranches, git.Branch{
				Name: branch,
				SHA:  branchToSHA[branch],
			})
		}
	}

	var fixCommits []string
	for hash, commit := range hashToCommit {
		if _, ok := hashToBranch[hash]; !ok {
			fixCommits = append(fixCommits, commit)
		}
	}

	if len(fixCommits) > 0 {
		return fixBranches, fixCommits, nil
	}

	for i := range fixBranches {
		err := g.DeleteBranch(fixBranches[i].Name)
		if err != nil {
			return nil, nil, err
		}
	}

	return nil, nil, nil
}

func gitCheckFeatureStack(ctx *cli.Context) error {
	fixBranches, fixCommits, err := check(ctx)
	if err != nil {
		return err
	}

	g := git.NewGit(ctx)

	if len(fixBranches) > 0 {
		fmt.Println("FIX BRANCHES")
		for _, branch := range fixBranches {
			commit, err := g.FindCommit(branch.Name)
			if err != nil {
				return err
			}
			fmt.Printf("git branch -f %s Subj:%s\n", branch.Name, commit.Subject)
		}
	}

	if len(fixCommits) > 0 {
		fmt.Println("FIX COMMITS")
		for _, sha := range fixCommits {
			commit, err := g.FindCommit(sha)
			if err != nil {
				return err
			}
			fmt.Printf("%s %s\n", commit.SHA, commit.Subject)
		}
	}

	return nil
}

func gitAssign(ctx *cli.Context) error {
	fixBranches, fixCommits, err := check(ctx)
	if err != nil {
		return err
	}

	g := git.NewGit(ctx)

	if ctx.NArg() < 2 {
		return errors.New("need branch and commit")
	}

	from, to := ctx.Args().First(), ctx.Args().Get(1)

	f := func() bool {
		for i := range fixBranches {
			if from == fixBranches[i].Name {
				return true
			}
		}
		return false
	}()

	if !f {
		return nil
	}

	x := func() bool {
		for i := range fixCommits {
			if to == fixCommits[i] {
				return true
			}
		}
		return false
	}()

	if !x {
		return nil
	}

	return g.SwitchBranch(from, to)
}

func reviewBranchPrefix(feature string) string {
	return fmt.Sprintf("review/%s/", feature)
}
