package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/waffleboot/giiter/internal/app"
)

var (
	cfgFile       string
	baseBranch    string
	featureBranch string
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return rootCmd.ExecuteContext(ctx)
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().BoolVar(&app.Config.Push, "push", false, "git push")
	rootCmd.PersistentFlags().StringVar(&app.Config.Repo, "repo", "", "path to git repository")
	rootCmd.PersistentFlags().BoolVarP(&app.Config.Debug, "debug", "d", false, "debug output")
	rootCmd.PersistentFlags().BoolVarP(&app.Config.Verbose, "verbose", "v", false, "verbose output")

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(makeCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(assignCmd)
	rootCmd.AddCommand(branchesCmd)

	listCmd.Flags().StringVarP(&baseBranch, "base", "b", "", "base branch")
	listCmd.Flags().StringVarP(&featureBranch, "feature", "f", "", "feature branch")
	listCmd.Flags().BoolVar(&app.Config.RefreshOnSubject, "refresh-on-subj", false, "refresh using by subject")

	makeCmd.Flags().StringVarP(&baseBranch, "base", "b", "", "base branch")
	makeCmd.Flags().StringVarP(&featureBranch, "feature", "f", "", "feature branch")
	makeCmd.Flags().StringVar(&app.Config.Prefix, "prefix", "", "merge review subj prefix")

	deleteCmd.Flags().StringVarP(&featureBranch, "feature", "f", "", "feature branch")

	assignCmd.Flags().StringVarP(&baseBranch, "base", "b", "", "base branch")
	assignCmd.Flags().StringVarP(&featureBranch, "feature", "f", "", "feature branch")
}

func initConfig() {
	if err := app.LoadConfig(cfgFile); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
