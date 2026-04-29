package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sage",
	Short: "🌿 Ask your system anything in plain English",
	Long: `
🌿 sage — Ask your system anything in plain English.
AI-powered system assistant for Linux & macOS.

Usage:
  sage "what's eating my memory"
  sage "why is my disk full"
  sage "what's running on port 3000"`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
