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

var _cfgFile string

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

	branches := new(branches)

	listCmd := makeListCommand(branches)
	makeCmd := makeMakeCommand(branches)
	diffCmd := makeDiffCommand(branches)
	assignCmd := makeAssignCommand(branches)
	rebaseCmd := makeRebaseCommand(branches)

	branches.addCommonFlags(listCmd)
	branches.addCommonFlags(makeCmd)
	branches.addCommonFlags(diffCmd)
	branches.addCommonFlags(rebaseCmd)
	branches.addCommonFlags(assignCmd)

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(makeCmd)
	rootCmd.AddCommand(assignCmd)
	rootCmd.AddCommand(rebaseCmd)
	rootCmd.AddCommand(makeDeleteCommand())
	rootCmd.AddCommand(makeBranchesCommand())
}

func initConfig() {
	if err := app.LoadConfig(_cfgFile); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
