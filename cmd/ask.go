package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yanonymousV2/sage/internal/ai"
	"github.com/yanonymousV2/sage/internal/executor"
)

func buildPrompt(question string) string {
	mem := executor.Run("vm_stat")
	disk := executor.Run("df -h /")
	procs := executor.Run("ps aux --sort=-%cpu | head -6")

	prompt := fmt.Sprintf(`You are sage, a system assistant for macOS and Linux terminals.
The user asked: "%s"

Here is their actual live system data:

DISK SPACE (main volume only):
%s

MEMORY:
%s

TOP PROCESSES BY CPU:
%s

Instructions:
- Read the numbers carefully and precisely
- For disk: the Avail column shows free space, Used shows used space
- Be specific with exact numbers from the output
- Be concise, plain English, no jargon
- If something needs fixing, tell the user exactly what command to run
- Keep response under 8 lines`, question, disk.Output, mem.Output, procs.Output)

	return prompt
}

var askCmd = &cobra.Command{
	Use:   "ask [question]",
	Short: "Ask your system a question in plain English",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		question := args[0]

		fmt.Println("🌿 Checking your system...")

		prompt := buildPrompt(question)

		response, err := ai.AskOllama("llama3.2", prompt)
		if err != nil {
			fmt.Println("ERR", err)
			return
		}

		fmt.Println("\n" + response)
	},
}

func init() {
	rootCmd.AddCommand(askCmd)
}
