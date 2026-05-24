package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const repoAPI = "https://api.github.com/repos/yanonymousV2/sage/releases/latest"

var updateClient = &http.Client{Timeout: 30 * time.Second}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update sage to the latest release",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()
		fmt.Println(stepStyle.Render("● checking for updates..."))

		// Fetch latest release info
		resp, err := updateClient.Get(repoAPI)
		if err != nil {
			fmt.Println(errorStyle.Render("❌ could not reach GitHub: " + err.Error()))
			return
		}
		defer resp.Body.Close()

		var release struct {
			TagName string `json:"tag_name"`
			Assets  []struct {
				Name               string `json:"name"`
				BrowserDownloadURL string `json:"browser_download_url"`
			} `json:"assets"`
		}
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &release); err != nil {
			fmt.Println(errorStyle.Render("❌ could not parse release info"))
			return
		}

		latest := release.TagName
		if latest == "" {
			fmt.Println(errorStyle.Render("❌ no releases found"))
			return
		}

		// Compare versions
		if Version != "dev" && Version == latest {
			fmt.Println(successStyle.Render("✓ already on latest: " + latest))
			fmt.Println()
			return
		}

		fmt.Println(stepStyle.Render(fmt.Sprintf("● updating %s → %s", Version, latest)))

		// Find the right asset for this OS/arch
		goos := runtime.GOOS
		goarch := runtime.GOARCH
		assetName := fmt.Sprintf("sage_%s_%s.tar.gz", goos, goarch)

		var downloadURL string
		for _, a := range release.Assets {
			if a.Name == assetName {
				downloadURL = a.BrowserDownloadURL
				break
			}
		}
		if downloadURL == "" {
			fmt.Println(errorStyle.Render(fmt.Sprintf("❌ no binary found for %s/%s", goos, goarch)))
			return
		}

		// Download tarball
		fmt.Println(stepStyle.Render("● downloading..."))
		dlResp, err := updateClient.Get(downloadURL)
		if err != nil {
			fmt.Println(errorStyle.Render("❌ download failed: " + err.Error()))
			return
		}
		defer dlResp.Body.Close()

		// Extract binary from tarball
		gz, err := gzip.NewReader(dlResp.Body)
		if err != nil {
			fmt.Println(errorStyle.Render("❌ could not read archive"))
			return
		}
		defer gz.Close()

		tr := tar.NewReader(gz)
		var binaryData []byte
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println(errorStyle.Render("❌ could not extract archive"))
				return
			}
			if hdr.Name == "sage" || strings.HasSuffix(hdr.Name, "/sage") {
				binaryData, err = io.ReadAll(tr)
				if err != nil {
					fmt.Println(errorStyle.Render("❌ could not read binary from archive"))
					return
				}
				break
			}
		}
		if len(binaryData) == 0 {
			fmt.Println(errorStyle.Render("❌ binary not found in archive"))
			return
		}

		// Find current executable path
		execPath, err := os.Executable()
		if err != nil {
			fmt.Println(errorStyle.Render("❌ could not find current binary path"))
			return
		}

		// Write to a temp file next to the binary, then rename (atomic replace)
		tmp := execPath + ".tmp"
		if err := os.WriteFile(tmp, binaryData, 0755); err != nil {
			fmt.Println(errorStyle.Render("❌ could not write update (try with sudo): " + err.Error()))
			_ = os.Remove(tmp)
			return
		}
		if err := os.Rename(tmp, execPath); err != nil {
			fmt.Println(errorStyle.Render("❌ could not replace binary (try with sudo): " + err.Error()))
			_ = os.Remove(tmp)
			return
		}

		fmt.Println(successStyle.Render("✓ updated to " + latest))
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
