package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yanonymousV2/sage/internal/config"
)

var (
	modelFlag      string
	resetTrustFlag bool
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or update sage configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()

		if modelFlag != "" {
			cfg.Model = modelFlag
			if err := config.Save(cfg); err != nil {
				fmt.Println(errorStyle.Render("❌ could not save config: " + err.Error()))
				return
			}
			fmt.Println(successStyle.Render("✓ model set to: " + cfg.Model))
			return
		}

		if resetTrustFlag {
			cfg.TrustReadOnly = false
			if err := config.Save(cfg); err != nil {
				fmt.Println(errorStyle.Render("❌ could not save config: " + err.Error()))
				return
			}
			fmt.Println(successStyle.Render("✓ command approval prompts re-enabled"))
			return
		}

		fmt.Println(labelStyle.Render("model:    ") + cfg.Model)
		if cfg.TrustReadOnly {
			fmt.Println(labelStyle.Render("approval: ") + "always trusted  " + dim("(reset with --reset-trust)"))
		} else {
			fmt.Println(labelStyle.Render("approval: ") + "ask every time")
		}
	},
}

func init() {
	configCmd.Flags().StringVar(&modelFlag, "model", "", "Set the Ollama model (e.g. qwen2.5:7b, llama3.2)")
	configCmd.Flags().BoolVar(&resetTrustFlag, "reset-trust", false, "Re-enable command approval prompts")
	rootCmd.AddCommand(configCmd)
}
