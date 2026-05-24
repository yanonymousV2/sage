package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/yanonymousV2/sage/internal/ai"
	"github.com/yanonymousV2/sage/internal/config"
	"github.com/yanonymousV2/sage/internal/executor"
	"github.com/yanonymousV2/sage/internal/history"
	"github.com/yanonymousV2/sage/internal/safety"
)

func buildFixPrompt(question string, results []executor.CommandResult, osName string) string {
	var sb strings.Builder
	for _, r := range results {
		if r.Error != "" {
			fmt.Fprintf(&sb, "$ %s (failed: %s)\n%s\n\n", r.Command, r.Error, r.Output)
		} else {
			fmt.Fprintf(&sb, "$ %s\n%s\n\n", r.Command, r.Output)
		}
	}

	return fmt.Sprintf(`You are sage, a system assistant for %s.

The user wants to fix: "%s"

Here are the results from running diagnostic commands:
%s
Do two things:

1. Explain the problem in plain English (2-4 sentences, no headers, no markdown, no bold, no backticks).

2. Then on a new line write exactly:
FIX: <single shell command to fix the problem>

If no fix is needed or possible, write:
FIX: none

Example:
Your disk is full because the docker overlay directory is taking up 18GB. Three large unused images account for most of the space.

FIX: docker system prune -af`, osName, question, sb.String())
}

var fixCmd = &cobra.Command{
	Use:   "fix [problem]",
	Short: "Diagnose a problem and apply a suggested fix",
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

		tw := getTermWidth()
		contentWidth := max(tw-6, 40)
		indent := "  "

		fmt.Println()
		fmt.Println(headerStyle.Render("🌿 sage fix"))
		fmt.Println()
		fmt.Println(stepStyle.Render("● diagnosing..."))

		commands, err := getCommands(question, osName, model, backend)
		if err != nil {
			fmt.Println(errorStyle.Render("❌ " + err.Error()))
			return
		}

		approved, err := filterAndApprove(commands)
		if err != nil {
			fmt.Println(stepStyle.Render("  cancelled."))
			return
		}

		fmt.Println()
		fmt.Println(stepStyle.Render("● running:"))
		var results []executor.CommandResult
		for _, c := range approved {
			fmt.Println(cmdListStyle.Render("→ " + c))
			results = append(results, executor.Run(c))
		}

		fmt.Println()
		fmt.Println(stepStyle.Render("● analyzing..."))
		fmt.Println()

		fmt.Println(lipgloss.NewStyle().Foreground(colorMuted).PaddingLeft(2).Render("problem"))
		fmt.Println(lipgloss.NewStyle().Foreground(colorPurple).Bold(true).PaddingLeft(2).Render(wordWrap(question, contentWidth)))
		fmt.Println()

		// Stream response, collecting full text to parse FIX: line
		sw := &streamWriter{width: contentWidth, indent: indent, lineLen: 0}
		fmt.Print(indent)

		prompt := buildFixPrompt(question, results, osName)
		var fullResponse strings.Builder

		_, err = backend.Stream(model, prompt, func(token string) {
			fullResponse.WriteString(token)
			// Only stream the explanation part (before FIX:)
			if !strings.Contains(fullResponse.String(), "\nFIX:") {
				sw.write(token)
			}
		})
		fmt.Println()
		fmt.Println()

		if err != nil {
			fmt.Println(errorStyle.Render("❌ " + err.Error()))
			return
		}

		response := fullResponse.String()

		// Parse explanation and fix command
		explanation := response
		fixCmd := ""

		if idx := strings.Index(response, "\nFIX:"); idx != -1 {
			explanation = strings.TrimSpace(response[:idx])
			fixLine := strings.TrimSpace(response[idx+5:]) // skip "\nFIX:"
			// take first line only
			if nl := strings.Index(fixLine, "\n"); nl != -1 {
				fixLine = fixLine[:nl]
			}
			fixLine = strings.TrimSpace(fixLine)
			if fixLine != "" && fixLine != "none" {
				fixCmd = fixLine
			}
		}

		// Save to history
		_ = history.Append(history.Entry{
			Question: "fix: " + question,
			Answer:   explanation,
			Provider: cfg.Provider,
			Model:    model,
			OS:       osName,
		})

		if fixCmd == "" {
			fmt.Println(stepStyle.Render("  no fix needed or no fix available."))
			fmt.Println()
			return
		}

		// Show suggested fix
		fmt.Println(sectionStyle.Render("  suggested fix:"))
		fmt.Println()
		fmt.Println(cmdListStyle.Render("  $ " + fixCmd))
		fmt.Println()

		// Warn if dangerous
		if safety.IsBlocked(fixCmd) {
			fmt.Println(blockedStyle.Render("  ✗ this command is blocked for safety reasons"))
			fmt.Println()
			return
		}
		if safety.IsDangerous(fixCmd) {
			fmt.Println(dangerStyle.Render("  ⚠ this command makes changes to your system"))
		}

		// Ask to apply
		fmt.Print(lipgloss.NewStyle().Foreground(colorMuted).PaddingLeft(2).Render("apply fix? [y/n]: "))
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return
		}
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if answer != "y" && answer != "yes" {
			fmt.Println(stepStyle.Render("  skipped."))
			fmt.Println()
			return
		}

		fmt.Println()
		fmt.Println(stepStyle.Render("● applying fix..."))
		result := executor.Run(fixCmd)
		if result.Error != "" {
			fmt.Println(errorStyle.Render("❌ " + result.Error))
		} else {
			fmt.Println(successStyle.Render("✓ done"))
			if strings.TrimSpace(result.Output) != "" {
				fmt.Println()
				fmt.Print(indent)
				fmt.Println(wordWrap(strings.TrimSpace(result.Output), contentWidth))
			}
		}
		fmt.Println()
	},
}

func init() {
	fixCmd.Flags().StringVarP(&modelOverrideFlag, "model", "m", "", "Override model for this query")
	rootCmd.AddCommand(fixCmd)
}
