package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yanonymousV2/sage/internal/ai"
	"github.com/yanonymousV2/sage/internal/executor"
)

var gradeFlag bool

type Grade struct {
	Confidence string   `json:"confidence"` // "high", "medium", "low"
	Issues     []string `json:"issues"`
}

func gradeAnswer(question string, results []executor.CommandResult, answer, osName, model string) (*Grade, error) {
	var sb strings.Builder
	for _, r := range results {
		output := strings.TrimSpace(r.Output)
		if len(output) > 300 {
			output = output[:300] + "..."
		}
		if r.Error != "" {
			fmt.Fprintf(&sb, "$ %s\nERROR: %s\n%s\n\n", r.Command, r.Error, output)
		} else {
			fmt.Fprintf(&sb, "$ %s\n%s\n\n", r.Command, output)
		}
	}

	prompt := fmt.Sprintf(`You are an accuracy evaluator for a system assistant on %s.

Question asked: "%s"

Commands run and their real output:
%s
Answer given to the user:
%s

Evaluate whether the answer accurately reflects the real command output above.
Consider: were the right commands used, does the answer correctly interpret the data, any factual errors or missing key info?

Respond with JSON only — no text outside the JSON:
{"confidence": "high|medium|low", "issues": ["describe specific issue"]}

Confidence:
- high: accurate and well supported by the data
- medium: mostly correct but has gaps or minor errors
- low: significantly inaccurate or not supported by the data

Use an empty issues array when there are no problems.`, osName, question, sb.String(), answer)

	response, err := ai.AskOllama(model, prompt)
	if err != nil {
		return nil, err
	}

	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 {
		return nil, fmt.Errorf("could not parse grade")
	}

	jsonStr := badJSONEscape.ReplaceAllString(response[start:end+1], `$1`)
	var grade Grade
	if err := json.Unmarshal([]byte(jsonStr), &grade); err != nil {
		return nil, fmt.Errorf("could not parse grade: %w", err)
	}
	if grade.Confidence == "" {
		grade.Confidence = "unknown"
	}

	return &grade, nil
}

func renderGrade(grade *Grade) {
	tw := getTermWidth()
	contentWidth := max(tw-12, 40)

	var dotStyle lipgloss.Style
	switch grade.Confidence {
	case "high":
		dotStyle = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	case "medium":
		dotStyle = lipgloss.NewStyle().Foreground(colorAmber).Bold(true)
	default:
		dotStyle = lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	}

	fmt.Println(
		lipgloss.NewStyle().Foreground(colorMuted).PaddingLeft(2).Render("  accuracy") +
			"  " + dotStyle.Render("● "+grade.Confidence),
	)
	for _, issue := range grade.Issues {
		fmt.Println(lipgloss.NewStyle().Foreground(colorMuted).PaddingLeft(4).
			Render("  · " + wordWrap(issue, contentWidth-8)))
	}
	fmt.Println()
}
