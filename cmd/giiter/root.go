package main

import (
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/waffleboot/giiter/internal/app"
)

func makeRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "giiter",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			f, err := os.OpenFile(_cfgFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
			if err != nil {
				return err
			}
			defer f.Close()

			enc := yaml.NewEncoder(f)
			if err := enc.Encode(app.Config.Persistent); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&_cfgFile, "config", ".giiter.yml", "config file")
	cmd.PersistentFlags().BoolVarP(&app.Config.Debug, "debug", "d", false, "debug output")
	cmd.PersistentFlags().BoolVarP(&app.Config.Verbose, "verbose", "v", false, "verbose output")
	cmd.PersistentFlags().BoolVarP(&app.Config.EnableGitPush, "push", "p", false, "enable git push")
	cmd.PersistentFlags().BoolVar(&app.Config.UseSubjectToMatch, "subj", false, "use commit subject to match")

	return cmd
}
