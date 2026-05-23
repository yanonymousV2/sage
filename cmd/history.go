package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/yanonymousV2/sage/internal/history"
)

var (
	historyLastFlag   bool
	historySearchFlag string
	historyClearFlag  bool
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Browse past questions and answers",
	Run: func(cmd *cobra.Command, args []string) {
		if historyClearFlag {
			if err := history.Clear(); err != nil {
				fmt.Println(errorStyle.Render("❌ could not clear history: " + err.Error()))
				return
			}
			fmt.Println(successStyle.Render("✓ history cleared"))
			return
		}

		entries, err := history.Load()
		if err != nil {
			fmt.Println(errorStyle.Render("❌ could not load history: " + err.Error()))
			return
		}
		if len(entries) == 0 {
			fmt.Println(stepStyle.Render("  no history yet — ask something first"))
			return
		}

		if historyLastFlag {
			e := entries[len(entries)-1]
			printEntry(e, true)
			return
		}

		if historySearchFlag != "" {
			needle := strings.ToLower(historySearchFlag)
			var matches []history.Entry
			for _, e := range entries {
				if strings.Contains(strings.ToLower(e.Question), needle) ||
					strings.Contains(strings.ToLower(e.Answer), needle) {
					matches = append(matches, e)
				}
			}
			if len(matches) == 0 {
				fmt.Println(stepStyle.Render("  no matches for: " + historySearchFlag))
				return
			}
			fmt.Println()
			for _, e := range matches {
				printEntry(e, false)
			}
			return
		}

		// Default: list all
		tw := getTermWidth()
		fmt.Println()
		for _, e := range entries {
			ts := lipgloss.NewStyle().Foreground(colorMuted).Render(e.CreatedAt.Format("Jan 02 15:04"))
			id := lipgloss.NewStyle().Foreground(colorPurple).Bold(true).Render(fmt.Sprintf("#%d", e.ID))
			q := wordWrap(e.Question, tw-20)
			fmt.Printf("  %s  %s  %s\n", id, ts, q)
		}
		fmt.Println()
		fmt.Println(stepStyle.Render("  use --last to show the full last answer, --search to filter"))
		fmt.Println()
	},
}

func printEntry(e history.Entry, full bool) {
	tw := getTermWidth()
	contentWidth := max(tw-6, 40)

	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(colorMuted).PaddingLeft(2).Render(
		fmt.Sprintf("#%d · %s · %s · %s", e.ID, e.CreatedAt.Format("Jan 02 15:04"), e.Provider, e.Model),
	))
	fmt.Println(lipgloss.NewStyle().Foreground(colorPurple).Bold(true).PaddingLeft(2).Render(
		wordWrap(e.Question, contentWidth),
	))
	fmt.Println()

	if full {
		indent := "  "
		for _, line := range strings.Split(wordWrap(e.Answer, contentWidth), "\n") {
			fmt.Println(indent + line)
		}
	}
	fmt.Println()
}

func init() {
	historyCmd.Flags().BoolVar(&historyLastFlag, "last", false, "Show the full last answer")
	historyCmd.Flags().StringVar(&historySearchFlag, "search", "", "Search history by keyword")
	historyCmd.Flags().BoolVar(&historyClearFlag, "clear", false, "Clear all history")
	rootCmd.AddCommand(historyCmd)
}
