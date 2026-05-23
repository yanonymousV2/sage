package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/yanonymousV2/sage/internal/ai"
	"github.com/yanonymousV2/sage/internal/config"
	"github.com/yanonymousV2/sage/internal/executor"
)

const testRuns = 3

var testCmd = &cobra.Command{
	Use:   "test [question]",
	Short: "Run a question 3x and check if the AI picks consistent commands",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		question := strings.Join(args, " ")
		cfg := config.Load()
		model := cfg.Model
		if modelOverrideFlag != "" {
			model = modelOverrideFlag
		}

		backend, err := ai.New(cfg.Provider, cfg.APIKey)
		if err != nil {
			fmt.Println(errorStyle.Render("❌ " + err.Error()))
			return
		}

		osInfo := executor.Run("uname -s")
		osName := resolveOS(strings.TrimSpace(osInfo.Output))

		fmt.Println()
		fmt.Println(headerStyle.Render("🌿 sage test"))
		fmt.Println()
		fmt.Println(
			lipgloss.NewStyle().Foreground(colorMuted).PaddingLeft(2).Render("question  ")+
				lipgloss.NewStyle().Foreground(colorPurple).Bold(true).Render(question),
		)
		fmt.Println()

		runs := make([][]string, testRuns)
		for i := range testRuns {
			fmt.Print(stepStyle.Render(fmt.Sprintf("● run %d  ", i+1)))
			cmds, err := getCommands(question, osName, model, backend)
			if err != nil {
				fmt.Println(errorStyle.Render("failed: " + err.Error()))
				return
			}
			runs[i] = cmds
			fmt.Println()
			for _, c := range cmds {
				fmt.Println(cmdListStyle.Render("  " + c))
			}
			fmt.Println()
		}

		// Compare all runs against run 1
		allMatch := true
		ref := sortedKey(runs[0])
		for _, run := range runs[1:] {
			if sortedKey(run) != ref {
				allMatch = false
				break
			}
		}

		if allMatch {
			fmt.Println(successStyle.Render(fmt.Sprintf("✓ consistent — all %d runs chose the same commands", testRuns)))
		} else {
			fmt.Println(errorStyle.Render("⚠ inconsistent — commands varied across runs"))
			fmt.Println(stepStyle.Render("  try a larger model: sage config --model qwen2.5:14b"))
		}
		fmt.Println()
	},
}

// sortedKey returns a canonical string for a command list, order-insensitive.
func sortedKey(cmds []string) string {
	sorted := make([]string, len(cmds))
	copy(sorted, cmds)
	sort.Strings(sorted)
	return strings.Join(sorted, "|")
}

func init() {
	testCmd.Flags().StringVarP(&modelOverrideFlag, "model", "m", "", "Override model for this test")
	rootCmd.AddCommand(testCmd)
}
