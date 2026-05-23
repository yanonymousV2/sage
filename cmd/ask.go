package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/yanonymousV2/sage/internal/ai"
	"github.com/yanonymousV2/sage/internal/executor"
)

var (
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	headerStyle = lipgloss.NewStyle().
			Foreground(special).
			Bold(true).
			PaddingLeft(1)

	responseStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			PaddingRight(2).
			PaddingTop(1).
			PaddingBottom(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(highlight)

	thinkingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Italic(true).
			PaddingLeft(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(highlight).
			Bold(true)
)

type CommandPlan struct {
	Commands []string `json:"commands"`
}

func getCommands(question string, osName string) ([]string, error) {
	prompt := fmt.Sprintf(`You are a system assistant for %s.
The user asked: "%s"

Return a JSON object with a list of safe, read-only shell commands to run to answer this question.
Only include commands that are safe and non-destructive.
Maximum 5 commands.

Example response:
{"commands": ["df -h /", "top -l 1 -s 0 | grep PhysMem"]}

Return only the JSON, nothing else.`, osName, question)

	response, err := ai.AskOllama("qwen2.5:7b", prompt)
	if err != nil {
		return nil, err
	}

	// extract JSON from response
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 {
		return nil, fmt.Errorf("could not parse commands from response")
	}

	jsonStr := response[start : end+1]
	var plan CommandPlan
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		return nil, fmt.Errorf("could not parse command plan: %w", err)
	}

	return plan.Commands, nil
}

func explainResults(question string, results map[string]string, osName string) (string, error) {
	var sb strings.Builder
	for cmd, output := range results {
		sb.WriteString(fmt.Sprintf("$ %s\n%s\n\n", cmd, output))
	}

	prompt := fmt.Sprintf(`You are sage, a system assistant for %s.

The user asked: "%s"

Here are the results from running relevant commands:
%s

Answer the user's question using the real data above.
Be concise, plain English, no jargon.
Format:
- One sentence verdict
- Bullet points with key numbers
- Only suggest a fix if something needs attention
- No markdown backticks`, osName, question, sb.String())

	return ai.AskOllama("qwen2.5:7b", prompt)
}

var askCmd = &cobra.Command{
	Use:   "ask [question]",
	Short: "Ask your system anything in plain English",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		question := strings.Join(args, " ")
		osInfo := executor.Run("uname -s")
		osName := strings.TrimSpace(osInfo.Output)

		fmt.Println()
		fmt.Println(headerStyle.Render("🌿 sage"))
		fmt.Println(thinkingStyle.Render("Figuring out what to check..."))

		commands, err := getCommands(question, osName)
		if err != nil {
			fmt.Println(labelStyle.Render("❌ " + err.Error()))
			return
		}

		fmt.Println(thinkingStyle.Render("Running " + fmt.Sprintf("%d", len(commands)) + " commands..."))
		fmt.Println()

		results := make(map[string]string)
		for _, c := range commands {
			fmt.Println(thinkingStyle.Render("  $ " + c))
			result := executor.Run(c)
			if result.Error == "" {
				results[c] = result.Output
			}
		}

		fmt.Println(thinkingStyle.Render("Analyzing results..."))

		response, err := explainResults(question, results, osName)
		if err != nil {
			fmt.Println(labelStyle.Render("❌ " + err.Error()))
			return
		}

		response = strings.ReplaceAll(response, "```", "")
		response = strings.TrimSpace(response)

		fmt.Println()
		fmt.Println(labelStyle.Render("  You asked: ") + question)
		fmt.Println()
		fmt.Println(responseStyle.Render(response))
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(askCmd)
}
