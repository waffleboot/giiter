package main

import (
	"github.com/spf13/cobra"
	"github.com/waffleboot/giiter/internal/app"
)

var rootCmd = &cobra.Command{
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if app.Config.Log != nil {
			app.Config.Log.Close()
		}
		return nil
	},
}

