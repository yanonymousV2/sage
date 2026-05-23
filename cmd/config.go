package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yanonymousV2/sage/internal/config"
)

var (
	modelFlag      string
	providerFlag   string
	apiKeyFlag     string
	resetTrustFlag bool
	resetAllFlag   bool
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or update sage configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()

		if resetAllFlag {
			if err := config.Save(config.Defaults()); err != nil {
				fmt.Println(errorStyle.Render("❌ could not reset config: " + err.Error()))
				return
			}
			fmt.Println(successStyle.Render("✓ config reset to defaults"))
			return
		}

		if providerFlag != "" {
			cfg.Provider = providerFlag
			// auto-set a sensible default model when switching provider
			if def, ok := config.DefaultModelFor[providerFlag]; ok {
				cfg.Model = def
			}
			if err := config.Save(cfg); err != nil {
				fmt.Println(errorStyle.Render("❌ could not save config: " + err.Error()))
				return
			}
			fmt.Println(successStyle.Render("✓ provider set to: " + cfg.Provider))
			fmt.Println(successStyle.Render("✓ model set to:    " + cfg.Model))
			return
		}

		if modelFlag != "" {
			cfg.Model = modelFlag
			if err := config.Save(cfg); err != nil {
				fmt.Println(errorStyle.Render("❌ could not save config: " + err.Error()))
				return
			}
			fmt.Println(successStyle.Render("✓ model set to: " + cfg.Model))
			return
		}

		if apiKeyFlag != "" {
			cfg.APIKey = apiKeyFlag
			if err := config.Save(cfg); err != nil {
				fmt.Println(errorStyle.Render("❌ could not save config: " + err.Error()))
				return
			}
			fmt.Println(successStyle.Render("✓ API key saved"))
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

		// Show current config
		fmt.Println(labelStyle.Render("provider: ") + cfg.Provider)
		fmt.Println(labelStyle.Render("model:    ") + cfg.Model)
		if cfg.APIKey != "" {
			fmt.Println(labelStyle.Render("api key:  ") + "set")
		}
		if cfg.TrustReadOnly {
			fmt.Println(labelStyle.Render("approval: ") + "always trusted  " + dim("(reset with --reset-trust)"))
		} else {
			fmt.Println(labelStyle.Render("approval: ") + "ask every time")
		}
	},
}

func init() {
	configCmd.Flags().StringVar(&providerFlag, "provider", "", "Set AI provider: ollama, claude, openai")
	configCmd.Flags().StringVar(&modelFlag, "model", "", "Set the model (e.g. qwen2.5:14b, claude-sonnet-4-6, gpt-4o-mini)")
	configCmd.Flags().StringVar(&apiKeyFlag, "api-key", "", "Set API key for Claude or OpenAI")
	configCmd.Flags().BoolVar(&resetTrustFlag, "reset-trust", false, "Re-enable command approval prompts")
	configCmd.Flags().BoolVar(&resetAllFlag, "reset", false, "Reset all config to defaults")
	rootCmd.AddCommand(configCmd)
}
