package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/waffleboot/giiter/internal/app"
	"github.com/waffleboot/giiter/internal/git"
)

var (
	_cfgFile       string
	_baseBranch    string
	_featureBranch string
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
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

	rootCmd.Use = "giiter"
	rootCmd.SilenceUsage = true

	rootCmd.PersistentFlags().StringVar(&_cfgFile, "config", ".giiter.yml", "config file")
	rootCmd.PersistentFlags().BoolVarP(&app.Config.Debug, "debug", "d", false, "debug output")
	rootCmd.PersistentFlags().BoolVarP(&app.Config.Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&app.Config.EnableGitPush, "push", "p", false, "enable git push")
	rootCmd.PersistentFlags().BoolVar(&app.Config.UseSubjectToMatch, "subj", false, "use commit subject to match")

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(cmdDiff)
	rootCmd.AddCommand(makeCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(assignCmd)
	rootCmd.AddCommand(rebaseCmd)
	rootCmd.AddCommand(branchesCmd)

	addCommonFlags := func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&_baseBranch, "base", "b", "", "base branch")
		cmd.Flags().StringVarP(&_featureBranch, "feature", "f", "", "feature branch")
		cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) (err error) {
			_baseBranch, _featureBranch, err = git.FindBaseAndFeatureBranches(cmd.Context(), _baseBranch, _featureBranch)

			return
		}
	}

	addCommonFlags(listCmd)
	addCommonFlags(makeCmd)
	addCommonFlags(cmdDiff)
	addCommonFlags(rebaseCmd)
	addCommonFlags(assignCmd)

	makeCmd.Flags().StringVarP(&app.Config.MergeRequestPrefix, "prefix", "t", "", "title prefix for merge request")

	deleteCmd.Flags().StringVarP(&_featureBranch, "feature", "f", "", "feature branch")
}

func initConfig() {
	if err := app.LoadConfig(_cfgFile); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
