package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/waffleboot/giiter/internal/app"
	"gopkg.in/yaml.v2"
)

var rootCmd = &cobra.Command{
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.OpenFile(cfgFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
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
