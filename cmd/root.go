package cmd

import (
	"fmt"
	"os"
	"strings"

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
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		runAsk(strings.Join(args, " "))
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&gradeFlag, "grade", "g", false, "Grade answer accuracy after the response")
	rootCmd.Flags().StringVarP(&modelOverrideFlag, "model", "m", "", "Override model for this query (e.g. llama3.2)")
}
