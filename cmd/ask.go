package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	xterm "github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
	"github.com/yanonymousV2/sage/internal/ai"
	"github.com/yanonymousV2/sage/internal/config"
	"github.com/yanonymousV2/sage/internal/executor"
)

// --- theme ---

var (
	colorGreen  = lipgloss.AdaptiveColor{Light: "#16A34A", Dark: "#4ADE80"}
	colorPurple = lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}
	colorMuted  = lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#71717A"}
	colorAmber  = lipgloss.AdaptiveColor{Light: "#D97706", Dark: "#FBB040"}
	colorRed    = lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#F87171"}
	colorBlue   = lipgloss.AdaptiveColor{Light: "#2563EB", Dark: "#93C5FD"}
)

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(colorGreen).Bold(true).PaddingLeft(2)

	stepStyle = lipgloss.NewStyle().
			Foreground(colorMuted).PaddingLeft(2)

	cmdListStyle = lipgloss.NewStyle().
			Foreground(colorBlue).PaddingLeft(4)

	sectionStyle = lipgloss.NewStyle().
			Foreground(colorPurple).Bold(true).PaddingLeft(2)

	dangerStyle = lipgloss.NewStyle().
			Foreground(colorAmber).PaddingLeft(4)

	blockedStyle = lipgloss.NewStyle().
			Foreground(colorRed).PaddingLeft(2)

	labelStyle = lipgloss.NewStyle().
			Foreground(colorPurple).Bold(true).PaddingLeft(2)

	successStyle = lipgloss.NewStyle().
			Foreground(colorGreen).PaddingLeft(2)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorRed).PaddingLeft(2)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPurple).
			PaddingLeft(3).PaddingRight(3).
			PaddingTop(1).PaddingBottom(1).
			MarginLeft(2)
)

// --- helpers ---

func getTermWidth() int {
	w, _, err := xterm.GetSize(os.Stdout.Fd())
	if err != nil || w <= 0 {
		return 80
	}
	if w > 120 {
		return 120
	}
	return w
}

// wordWrap wraps text to fit within width cells, respecting existing newlines
// and using lipgloss.Width for correct Unicode/CJK character counting.
func wordWrap(text string, width int) string {
	if width <= 0 {
		return text
	}
	var out []string
	for line := range strings.SplitSeq(text, "\n") {
		if lipgloss.Width(line) <= width {
			out = append(out, line)
			continue
		}
		var cur strings.Builder
		curLen := 0
		for i, word := range strings.Fields(line) {
			wLen := lipgloss.Width(word)
			if i > 0 && curLen+1+wLen > width {
				out = append(out, cur.String())
				cur.Reset()
				curLen = 0
			}
			if curLen > 0 {
				cur.WriteByte(' ')
				curLen++
			}
			cur.WriteString(word)
			curLen += wLen
		}
		if cur.Len() > 0 {
			out = append(out, cur.String())
		}
	}
	return strings.Join(out, "\n")
}

// badJSONEscape matches backslashes followed by characters that are not valid
// JSON escape sequences. Shell commands like `find -exec ... \;` produce these.
var badJSONEscape = regexp.MustCompile(`\\([^"\\\/bfnrtu])`)

// --- logic ---

type CommandPlan struct {
	Commands []string `json:"commands"`
}

func getCommands(question, osName, model string) ([]string, error) {
	prompt := fmt.Sprintf(`You are a system assistant for %s.
The user asked: "%s"

Return a JSON object with a list of safe, read-only shell commands to answer this question.
Rules:
- Only safe, non-destructive commands
- No backslashes in commands (avoid find -exec, use xargs instead)
- Maximum 5 commands
- Return only raw JSON, no markdown, no explanation

Example: {"commands": ["df -h /", "ps aux | grep postgres"]}`, osName, question)

	response, err := ai.AskOllama(model, prompt)
	if err != nil {
		return nil, err
	}

	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 {
		return nil, fmt.Errorf("could not parse commands from response")
	}

	jsonStr := badJSONEscape.ReplaceAllString(response[start:end+1], `$1`)

	var plan CommandPlan
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		return nil, fmt.Errorf("could not parse command plan: %w", err)
	}

	return plan.Commands, nil
}

func explainResults(question string, results []executor.CommandResult, osName, model string) (string, error) {
	var sb strings.Builder
	for _, r := range results {
		if r.Error != "" {
			fmt.Fprintf(&sb, "$ %s (failed: %s)\n%s\n\n", r.Command, r.Error, r.Output)
		} else {
			fmt.Fprintf(&sb, "$ %s\n%s\n\n", r.Command, r.Output)
		}
	}

	prompt := fmt.Sprintf(`You are sage, a system assistant for %s.

The user asked: "%s"

Here are the results from running relevant commands:
%s
Answer in plain English using the real data above.
Rules:
- No section headers or labels (no "Verdict:", "Key Numbers:", "Fix Suggested:" etc.)
- Start with one clear sentence summarising the situation
- Follow with short bullet points for the key facts and numbers
- If something needs a fix, add one short suggestion at the end
- No markdown, no backticks, respond in the same language as the question`, osName, question, sb.String())

	return ai.AskOllama(model, prompt)
}

func runAsk(question string) {
	cfg := config.Load()
	osInfo := executor.Run("uname -s")
	osName := strings.TrimSpace(osInfo.Output)

	fmt.Println()
	fmt.Println(headerStyle.Render("🌿 sage"))
	fmt.Println()
	fmt.Println(stepStyle.Render("● figuring out what to check..."))

	commands, err := getCommands(question, osName, cfg.Model)
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

	response, err := explainResults(question, results, osName, cfg.Model)
	if err != nil {
		fmt.Println(errorStyle.Render("❌ " + err.Error()))
		return
	}

	response = strings.ReplaceAll(response, "```", "")
	response = strings.TrimSpace(response)

	// Fit box to terminal: margin(2) + border(2) + padding(6) + right buffer(2) = 12
	tw := getTermWidth()
	contentWidth := tw - 12
	contentWidth = max(contentWidth, 40)

	qLabel := lipgloss.NewStyle().Foreground(colorMuted).Render("you asked")
	qText := lipgloss.NewStyle().Foreground(colorPurple).Bold(true).Render(wordWrap(question, contentWidth))
	body := qLabel + "\n" + qText + "\n\n" + wordWrap(response, contentWidth)

	fmt.Println()
	fmt.Println(boxStyle.Width(contentWidth).Render(body))
	fmt.Println()

	if gradeFlag {
		fmt.Println(stepStyle.Render("● grading answer..."))
		grade, err := gradeAnswer(question, results, response, osName, cfg.Model)
		if err != nil {
			fmt.Println(errorStyle.Render("  (grade failed: " + err.Error() + ")"))
		} else {
			renderGrade(grade)
		}
	}
}

var askCmd = &cobra.Command{
	Use:   "ask [question]",
	Short: "Ask your system anything in plain English",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runAsk(strings.Join(args, " "))
	},
}

func init() {
	rootCmd.AddCommand(askCmd)
}
