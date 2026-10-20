package release

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"roblox-sync/sync-agent/internal/packageplugin"
)

const (
	WindowsAsset = "rbx-sync-windows-amd64.exe"
	LinuxAsset   = "rbx-sync-linux-amd64"
	PluginAsset  = "rbx-sync.rbxmx"
	SumsName     = "SHA256SUMS"
)

func goCmd() string {
	if p, err := exec.LookPath("go"); err == nil {
		return p
	}
	return "go"
}

func CrossCompile(agentDir, windowsOut, linuxOut string) error {
	if err := build(agentDir, "windows", "amd64", windowsOut); err != nil {
		return err
	}
	return build(agentDir, "linux", "amd64", linuxOut)
}

func build(agentDir, goos, goarch, out string) error {
	cmd := exec.Command(goCmd(), "build", "-o", out, "./cmd/rbx-sync")
	cmd.Dir = agentDir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS="+goos, "GOARCH="+goarch)
	if b, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go build GOOS=%s GOARCH=%s: %w\n%s", goos, goarch, err, b)
	}
	return nil
}

func Package(agentDir, pluginSrc, outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	win := filepath.Join(outDir, WindowsAsset)
	linux := filepath.Join(outDir, LinuxAsset)
	plugin := filepath.Join(outDir, PluginAsset)
	if err := CrossCompile(agentDir, win, linux); err != nil {
		return err
	}
	if err := packageplugin.Package(pluginSrc, plugin); err != nil {
		return err
	}
	return writeSums(outDir, WindowsAsset, LinuxAsset, PluginAsset)
}

func writeSums(outDir string, names ...string) error {
	var b strings.Builder
	for _, name := range names {
		path := filepath.Join(outDir, name)
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(body)
		fmt.Fprintf(&b, "%s  %s\n", hex.EncodeToString(sum[:]), name)
	}
	return os.WriteFile(filepath.Join(outDir, SumsName), []byte(b.String()), 0o644)
}
