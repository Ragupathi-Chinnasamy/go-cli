package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "Go-cli",
	Aliases: []string{"go-cli", "cli"},
	Short:   "Go-cli is a tool for setting up initial files for a Go backend project",
	Long:    "Go-cli is a tool for setting up initial files for a Go backend project which includes go.mod, main.go, .env, routes, config, services, repositories, models and utils",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Oops. An error while executing go-cli '%s'\n", err)
		os.Exit(1)
	}
}
