package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/yanonymousV2/sage/internal/config"
)

var (
	modelFlag       string
	providerFlag    string
	apiKeyFlag      string
	resetTrustFlag  bool
	resetAllFlag    bool
	listModelsFlag  bool
)

func listOllamaModels() {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://localhost:11434/api/tags")
	if err != nil {
		fmt.Println(errorStyle.Render("❌ could not reach Ollama — is it running? Start with: ollama serve"))
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Models []struct {
			Name string `json:"name"`
			Size int64  `json:"size"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &result); err != nil || len(result.Models) == 0 {
		fmt.Println(stepStyle.Render("  no models found — pull one with: ollama pull qwen2.5:14b"))
		return
	}

	cfg := config.Load()
	fmt.Println()
	for _, m := range result.Models {
		active := "  "
		style := cmdListStyle
		if m.Name == cfg.Model {
			active = "✓ "
			style = successStyle
		}
		sizeMB := m.Size / 1024 / 1024
		fmt.Println(style.Render(fmt.Sprintf("%s%-30s %d MB", active, m.Name, sizeMB)))
	}
	fmt.Println()
	fmt.Println(stepStyle.Render("  switch model: sage config --model <name>"))
	fmt.Println()
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or update sage configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()

		if listModelsFlag {
			listOllamaModels()
			return
		}

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
	configCmd.Flags().BoolVar(&listModelsFlag, "list-models", false, "List locally available Ollama models")
	rootCmd.AddCommand(configCmd)
}
