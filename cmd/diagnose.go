package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/yanonymousV2/sage/internal/ai"
	"github.com/yanonymousV2/sage/internal/config"
	"github.com/yanonymousV2/sage/internal/executor"
)

type diagSection struct {
	name     string
	commands []string
}

var diagSections = []diagSection{
	{
		name: "System",
		commands: []string{
			"uname -sr",
			"hostname",
			"uptime",
		},
	},
	{
		name: "CPU & Memory",
		commands: []string{
			"top -l 1 -s 0 | head -12",
			"sysctl -n hw.logicalcpu hw.memsize 2>/dev/null || nproc && free -h",
		},
	},
	{
		name: "Disk",
		commands: []string{
			"df -h | grep -v tmpfs",
		},
	},
	{
		name: "Network",
		commands: []string{
			"ifconfig 2>/dev/null | grep 'inet ' || ip addr show | grep 'inet '",
			"netstat -an 2>/dev/null | grep LISTEN | head -10",
		},
	},
	{
		name: "Top Processes",
		commands: []string{
			"ps aux --sort=-%cpu 2>/dev/null | head -8 || ps aux | head -8",
		},
	},
}

func buildDiagPrompt(sectionName string, results []executor.CommandResult, osName string) string {
	var sb strings.Builder
	for _, r := range results {
		if r.Error != "" {
			fmt.Fprintf(&sb, "$ %s\nERROR: %s\n\n", r.Command, r.Error)
		} else {
			fmt.Fprintf(&sb, "$ %s\n%s\n\n", r.Command, strings.TrimSpace(r.Output))
		}
	}

	return fmt.Sprintf(`You are sage, a system health monitor for %s.

Section: %s

Real system data:
%s
Give a 1-2 sentence health summary for this section only.
Be specific with numbers. Flag anything abnormal.
Plain text only — no markdown, no bold, no bullets, no labels.`, osName, sectionName, sb.String())
}

var diagnoseCmd = &cobra.Command{
	Use:   "diagnose",
	Short: "Run a full system health report",
	Run: func(cmd *cobra.Command, args []string) {
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

		tw := getTermWidth()
		contentWidth := max(tw-8, 40)

		fmt.Println()
		fmt.Println(headerStyle.Render("🌿 sage diagnose"))
		fmt.Println(lipgloss.NewStyle().Foreground(colorMuted).PaddingLeft(2).Render(
			osName + "  ·  " + time.Now().Format("Jan 02 15:04"),
		))
		fmt.Println()

		for _, section := range diagSections {
			// Run commands for this section
			var results []executor.CommandResult
			for _, c := range section.commands {
				results = append(results, executor.Run(c))
			}

			// Section header
			fmt.Println(sectionStyle.Render("  " + section.name))
			fmt.Println()

			// Get AI summary for this section
			prompt := buildDiagPrompt(section.name, results, osName)
			summary, err := backend.Complete(model, prompt)
			if err != nil {
				fmt.Println(errorStyle.Render("  ❌ " + err.Error()))
				fmt.Println()
				continue
			}

			summary = strings.TrimSpace(summary)
			indent := "    "
			for _, line := range strings.Split(wordWrap(summary, contentWidth), "\n") {
				fmt.Println(indent + line)
			}
			fmt.Println()
		}

		fmt.Println(lipgloss.NewStyle().Foreground(colorMuted).PaddingLeft(2).Render(
			"run 'sage fix <problem>' to address any issues",
		))
		fmt.Println()
	},
}

func init() {
	diagnoseCmd.Flags().StringVarP(&modelOverrideFlag, "model", "m", "", "Override model for this diagnose")
	rootCmd.AddCommand(diagnoseCmd)
}
