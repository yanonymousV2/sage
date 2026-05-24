package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/yanonymousV2/sage/internal/ai"
	"github.com/yanonymousV2/sage/internal/config"
	"github.com/yanonymousV2/sage/internal/executor"
)

var watchInterval int

func buildWatchPrompt(question string, results []executor.CommandResult, osName string) string {
	var sb strings.Builder
	for _, r := range results {
		if r.Error != "" {
			fmt.Fprintf(&sb, "$ %s (failed: %s)\n%s\n\n", r.Command, r.Error, r.Output)
		} else {
			fmt.Fprintf(&sb, "$ %s\n%s\n\n", r.Command, r.Output)
		}
	}

	return fmt.Sprintf(`You are sage, a system monitor for %s.

The user is watching: "%s"

Real system data:
%s
Reply in ONE short sentence (max 15 words) summarising the current status.
No markdown, no labels, no punctuation at the end.
Example: "CPU is at 34%% across 8 cores, all normal"`, osName, question, sb.String())
}

var watchCmd = &cobra.Command{
	Use:   "watch [question]",
	Short: "Monitor your system and alert when something changes",
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

		// Resolve commands once — reuse every tick
		fmt.Println()
		fmt.Println(headerStyle.Render("🌿 sage watch"))
		fmt.Println(lipgloss.NewStyle().Foreground(colorMuted).PaddingLeft(2).Render(
			fmt.Sprintf("watching every %ds  ·  Ctrl+C to stop", watchInterval),
		))
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Foreground(colorPurple).Bold(true).PaddingLeft(2).Render(question))
		fmt.Println()

		fmt.Print(stepStyle.Render("● planning commands..."))
		commands, err := getCommands(question, osName, model, backend)
		if err != nil {
			fmt.Println(errorStyle.Render(" ❌ " + err.Error()))
			return
		}
		fmt.Println()
		for _, c := range commands {
			fmt.Println(cmdListStyle.Render("  → " + c))
		}
		fmt.Println()

		// Handle Ctrl+C gracefully
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

		ticker := time.NewTicker(time.Duration(watchInterval) * time.Second)
		defer ticker.Stop()

		var lastStatus string
		tick := func() {
			var results []executor.CommandResult
			for _, c := range commands {
				results = append(results, executor.Run(c))
			}

			prompt := buildWatchPrompt(question, results, osName)
			status, err := backend.Complete(model, prompt)
			if err != nil {
				ts := lipgloss.NewStyle().Foreground(colorMuted).Render(time.Now().Format("15:04:05"))
				fmt.Printf("  %s  %s\n", ts, errorStyle.Render("error: "+err.Error()))
				return
			}
			status = strings.TrimSpace(status)

			ts := lipgloss.NewStyle().Foreground(colorMuted).Render(time.Now().Format("15:04:05"))

			if lastStatus == "" {
				// First tick
				fmt.Printf("  %s  %s\n", ts,
					lipgloss.NewStyle().Foreground(colorGreen).Render(status))
			} else if !statusEqual(status, lastStatus) {
				// Status changed — highlight
				fmt.Printf("  %s  %s  %s\n", ts,
					lipgloss.NewStyle().Foreground(colorAmber).Bold(true).Render("⚠ changed:"),
					lipgloss.NewStyle().Foreground(colorAmber).Render(status))
			} else {
				// No change
				fmt.Printf("  %s  %s\n", ts,
					lipgloss.NewStyle().Foreground(colorMuted).Render(status))
			}

			lastStatus = status
		}

		// Run immediately on start
		tick()

		for {
			select {
			case <-ticker.C:
				tick()
			case <-sig:
				fmt.Println()
				fmt.Println(stepStyle.Render("  stopped."))
				fmt.Println()
				return
			}
		}
	},
}

// statusEqual does a loose comparison ignoring minor wording differences.
func statusEqual(a, b string) bool {
	normalize := func(s string) string {
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, ",", "")
		s = strings.ReplaceAll(s, ".", "")
		return strings.TrimSpace(s)
	}
	return normalize(a) == normalize(b)
}

func init() {
	watchCmd.Flags().IntVarP(&watchInterval, "interval", "i", 30, "Polling interval in seconds")
	watchCmd.Flags().StringVarP(&modelOverrideFlag, "model", "m", "", "Override model for this watch")
	rootCmd.AddCommand(watchCmd)
}
