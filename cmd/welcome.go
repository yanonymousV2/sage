package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yanonymousV2/sage/internal/config"
	"github.com/yanonymousV2/sage/internal/executor"
)

func showWelcome() {
	cfg := config.Load()
	osInfo := executor.Run("uname -s")
	osName := resolveOS(strings.TrimSpace(osInfo.Output))

	fmt.Println()
	fmt.Println(
		headerStyle.Render("🌿 sage") +
			lipgloss.NewStyle().Foreground(colorMuted).Render("  "+Version+"  ·  ask your system anything"),
	)
	fmt.Println()

	fmt.Println(sectionStyle.Render("  try asking:"))
	fmt.Println()
	for _, e := range []string{
		`sage "why is my disk full"`,
		`sage "what's using port 3000"`,
		`sage "is postgres running"`,
		`sage "how is my memory looking"`,
	} {
		fmt.Println(cmdListStyle.Render(e))
	}

	fmt.Println()
	fmt.Println(sectionStyle.Render("  commands:"))
	fmt.Println()

	rows := [][2]string{
		{"sage <question>", "ask anything"},
		{"sage config", "show current settings"},
		{"sage config --model <name>", "switch AI model"},
		{"sage history", "browse past questions"},
		{"sage test <question>", "check answer consistency"},
		{"sage --grade <question>", "grade answer accuracy"},
	}
	for _, row := range rows {
		label := cmdListStyle.Render(fmt.Sprintf("%-32s", row[0]))
		desc := lipgloss.NewStyle().Foreground(colorMuted).Render(row[1])
		fmt.Println(label + desc)
	}

	fmt.Println()
	fmt.Println(
		lipgloss.NewStyle().Foreground(colorMuted).PaddingLeft(4).Render("model  ")+
			lipgloss.NewStyle().Foreground(colorPurple).Bold(true).Render(cfg.Model)+
			lipgloss.NewStyle().Foreground(colorMuted).Render("  ·  "+osName),
	)
	fmt.Println()
}
