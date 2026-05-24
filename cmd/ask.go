package cmd

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	xterm "github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
	"github.com/yanonymousV2/sage/internal/ai"
	sagecontext "github.com/yanonymousV2/sage/internal/context"
	"github.com/yanonymousV2/sage/internal/config"
	"github.com/yanonymousV2/sage/internal/executor"
	"github.com/yanonymousV2/sage/internal/history"
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
)

var modelOverrideFlag string
var followUpFlag bool

// badJSONEscape matches backslashes followed by characters that are not valid
// JSON escape sequences. Shell commands like `find -exec ... \;` produce these.
var badJSONEscape = regexp.MustCompile(`\\([^"\\\/bfnrtu])`)

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

// streamWriter prints tokens to the terminal with word-wrapping and indentation.
type streamWriter struct {
	width   int
	indent  string
	lineLen int
}

func (sw *streamWriter) write(token string) {
	for _, r := range token {
		switch r {
		case '\n':
			fmt.Println()
			fmt.Print(sw.indent)
			sw.lineLen = 0
		case ' ':
			if sw.lineLen > 0 && sw.lineLen >= sw.width {
				fmt.Println()
				fmt.Print(sw.indent)
				sw.lineLen = 0
			} else {
				fmt.Print(" ")
				sw.lineLen++
			}
		default:
			fmt.Printf("%c", r)
			sw.lineLen += lipgloss.Width(string(r))
		}
	}
}

// readStdin returns piped stdin content, or "" if stdin is a terminal.
func readStdin() string {
	stat, err := os.Stdin.Stat()
	if err != nil || (stat.Mode()&os.ModeCharDevice) != 0 {
		return "" // interactive terminal, nothing piped
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// resolveOS maps raw uname output to a human-readable OS name the model understands.
func resolveOS(uname string) string {
	switch uname {
	case "Darwin":
		return "macOS"
	case "Linux":
		return "Linux"
	default:
		return uname
	}
}

// --- logic ---

func osToolHint(osName string) string {
	switch osName {
	case "macOS":
		return "Use macOS commands: vm_stat, top -l 1 -s 0, sysctl, sw_vers, system_profiler, lsof, netstat, launchctl, brew. Do NOT use Linux-only commands (free, vmstat, /proc/*, apt, yum, systemctl)."
	default:
		return "Use Linux commands: free, vmstat, /proc/meminfo, lsb_release, systemctl, apt, journalctl."
	}
}

func getCommands(question, osName, model string, backend ai.Backend) ([]string, error) {
	prompt := fmt.Sprintf(`You are a system assistant for %s.
The user asked: "%s"

%s

List up to 5 safe, read-only shell commands to answer this question.
Output ONLY the commands, one per line, nothing else.
No explanations, no numbers, no bullets, no JSON, no markdown.

Example output:
df -h /
ps aux | grep postgres
lsof -i :3000`, osName, question, osToolHint(osName))

	response, err := backend.Complete(model, prompt)
	if err != nil {
		return nil, err
	}

	var commands []string
	for line := range strings.SplitSeq(strings.TrimSpace(response), "\n") {
		line = strings.TrimSpace(line)
		// skip blank lines and markdown fences
		if line == "" || line == "```" || strings.HasPrefix(line, "```") {
			continue
		}
		// strip leading markdown list markers (-, *, •, 1., 2. …)
		line = strings.TrimLeft(line, "-*•0123456789. \t")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		commands = append(commands, line)
		if len(commands) >= 5 {
			break
		}
	}

	if len(commands) == 0 {
		return nil, fmt.Errorf("no commands returned from model")
	}

	return commands, nil
}

func buildExplainPrompt(question string, results []executor.CommandResult, osName, pipeData string) string {
	var sb strings.Builder
	for _, r := range results {
		if r.Error != "" {
			fmt.Fprintf(&sb, "$ %s (failed: %s)\n%s\n\n", r.Command, r.Error, r.Output)
		} else {
			fmt.Fprintf(&sb, "$ %s\n%s\n\n", r.Command, r.Output)
		}
	}

	pipedSection := ""
	if pipeData != "" {
		// truncate very large stdin to avoid blowing the context window
		if len(pipeData) > 4000 {
			pipeData = pipeData[:4000] + "\n... (truncated)"
		}
		pipedSection = fmt.Sprintf("\nThe user also piped in this data:\n%s\n", pipeData)
	}

	return fmt.Sprintf(`You are sage, a system assistant for %s.

The user asked: "%s"
%s
Here are the results from running relevant commands:
%s
Answer in plain English using the real data above.
Rules:
- No section headers or labels (no "Verdict:", "Key Numbers:", "Fix Suggested:" etc.)
- Start with one clear sentence summarising the situation
- Follow with short bullet points for the key facts and numbers
- If something needs a fix, add one short suggestion at the end
- Plain text only — no markdown, no bold (**), no italics (_), no backticks, no bullet symbols like •
- Use a simple dash - for bullet points
- Respond in the same language as the question`, osName, question, pipedSection, sb.String())
}

func runAsk(question string) {
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

	pipeData := readStdin()

	osInfo := executor.Run("uname -s")
	osName := resolveOS(strings.TrimSpace(osInfo.Output))

	tw := getTermWidth()
	contentWidth := max(tw-6, 40)
	indent := "  "

	fmt.Println()
	fmt.Println(headerStyle.Render("🌿 sage"))
	fmt.Println()
	fmt.Println(stepStyle.Render("● figuring out what to check..."))

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

	// Show question header before streaming starts
	label := "you asked"
	if priorSession != nil {
		label = "follow-up"
	}
	fmt.Println(lipgloss.NewStyle().Foreground(colorMuted).PaddingLeft(2).Render(label))
	fmt.Println(lipgloss.NewStyle().Foreground(colorPurple).Bold(true).PaddingLeft(2).Render(wordWrap(question, contentWidth)))
	fmt.Println()

	// Stream the response live
	sw := &streamWriter{width: contentWidth, indent: indent, lineLen: 0}
	fmt.Print(indent)

	prompt := buildExplainPrompt(question, results, osName, pipeData)
	response, err := backend.Stream(model, prompt, func(token string) {
		sw.write(token)
	})
	fmt.Println()
	fmt.Println()

	if err != nil {
		fmt.Println(errorStyle.Render("❌ " + err.Error()))
		return
	}

	response = strings.TrimSpace(strings.ReplaceAll(response, "```", ""))

	if gradeFlag {
		fmt.Println(stepStyle.Render("● grading answer..."))
		grade, err := gradeAnswer(question, results, response, osName, model, backend)
		if err != nil {
			fmt.Println(errorStyle.Render("  (grade failed: " + err.Error() + ")"))
		} else {
			renderGrade(grade)
		}
	}

	_ = history.Append(history.Entry{
		Question: question,
		Answer:   response,
		Provider: cfg.Provider,
		Model:    model,
		OS:       osName,
	})
	_ = sagecontext.Save(sagecontext.Session{
		Question: question,
		Answer:   response,
		OS:       osName,
	})
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
	askCmd.Flags().StringVarP(&modelOverrideFlag, "model", "m", "", "Override model for this query (e.g. llama3.2)")
	askCmd.Flags().BoolVarP(&followUpFlag, "follow-up", "f", false, "Include previous Q&A as context")
	rootCmd.AddCommand(askCmd)
}
