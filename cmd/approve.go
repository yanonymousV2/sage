package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yanonymousV2/sage/internal/config"
	"github.com/yanonymousV2/sage/internal/safety"
)

// sessionTrustReadOnly is set when user picks [s]ession — lasts until process exits.
var sessionTrustReadOnly bool

// key renders a bracketed key in green bold, e.g. [y]
func key(k string) string {
	return lipgloss.NewStyle().Foreground(colorGreen).Bold(true).Render("["+k+"]")
}

// dim renders text in muted color
func dim(s string) string {
	return lipgloss.NewStyle().Foreground(colorMuted).Render(s)
}

// filterAndApprove removes blocked commands, warns about dangerous ones,
// and asks the user to approve the rest before anything runs.
func filterAndApprove(commands []string) ([]string, error) {
	var blocked, dangerous, safe []string

	for _, cmd := range commands {
		switch {
		case safety.IsBlocked(cmd):
			blocked = append(blocked, cmd)
		case safety.IsDangerous(cmd):
			dangerous = append(dangerous, cmd)
		default:
			safe = append(safe, cmd)
		}
	}

	if len(blocked) > 0 {
		fmt.Println()
		for _, c := range blocked {
			fmt.Println(blockedStyle.Render("❌ blocked: " + c))
		}
	}

	runnable := append(safe, dangerous...)
	if len(runnable) == 0 {
		return nil, fmt.Errorf("no safe commands to run")
	}

	// Auto-approve if user already trusts read-only commands.
	cfg := config.Load()
	if len(dangerous) == 0 && (sessionTrustReadOnly || cfg.TrustReadOnly) {
		return runnable, nil
	}

	// Show what will run.
	fmt.Println()
	fmt.Println(sectionStyle.Render("sage wants to run:"))
	fmt.Println()
	for _, c := range safe {
		fmt.Println(cmdListStyle.Render("$ " + c))
	}
	for _, c := range dangerous {
		fmt.Println(dangerStyle.Render("⚠  $ " + c + "  ← dangerous"))
	}
	fmt.Println()

	// Prompt with inline highlighted keys.
	cursor := lipgloss.NewStyle().Foreground(colorPurple).Render("› ")
	if len(dangerous) > 0 {
		fmt.Print(
			dim("  allow?  ") +
				key("y") + dim(" once   ") +
				key("n") + dim(" cancel   ") +
				cursor,
		)
	} else {
		fmt.Print(
			dim("  allow?  ") +
				key("y") + dim(" once   ") +
				key("s") + dim(" session   ") +
				key("a") + dim(" always   ") +
				key("n") + dim(" cancel   ") +
				cursor,
		)
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	choice := strings.ToLower(strings.TrimSpace(scanner.Text()))
	fmt.Println()

	switch choice {
	case "y", "yes":
		return runnable, nil

	case "s", "session":
		if len(dangerous) > 0 {
			return nil, fmt.Errorf("cancelled")
		}
		sessionTrustReadOnly = true
		return safe, nil

	case "a", "always":
		if len(dangerous) > 0 {
			return nil, fmt.Errorf("cancelled")
		}
		cfg.TrustReadOnly = true
		if err := config.Save(cfg); err != nil {
			fmt.Println(stepStyle.Render("  (could not save preference: " + err.Error() + ")"))
		}
		sessionTrustReadOnly = true
		return safe, nil

	default:
		return nil, fmt.Errorf("cancelled")
	}
}
