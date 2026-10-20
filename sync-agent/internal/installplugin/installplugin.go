package installplugin

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	pluginFileName     = "rbx-sync.rbxmx"
	expectedWindowsDir = `%LOCALAPPDATA%\Roblox\Plugins`
	mntPluginsGlob     = "/mnt/c/Users/*/AppData/Local/Roblox/Plugins"
)

// Env is the injectable environment for destination and --file defaults.
type Env struct {
	GOOS           string
	LocalAppData   string
	Home           string
	ExecutableDir  string
	LookPath       func(file string) (string, error)
	CombinedOutput func(name string, args ...string) ([]byte, error)
	Glob           func(pattern string) ([]string, error)
}

func DefaultEnv() Env {
	exe, err := os.Executable()
	dir := ""
	if err == nil {
		dir = filepath.Dir(exe)
	}
	home, _ := os.UserHomeDir()
	return Env{
		GOOS:          runtime.GOOS,
		LocalAppData:  os.Getenv("LOCALAPPDATA"),
		Home:          home,
		ExecutableDir: dir,
		LookPath:      exec.LookPath,
		CombinedOutput: func(name string, args ...string) ([]byte, error) {
			return exec.Command(name, args...).CombinedOutput()
		},
		Glob: filepath.Glob,
	}
}

func Command(env Env, args []string, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "install" {
		name := "plugin"
		if len(args) > 0 {
			name = args[0]
		}
		fmt.Fprintf(stderr, "unknown command: %s\n", name)
		return 2
	}
	fs := flag.NewFlagSet("plugin install", flag.ContinueOnError)
	fs.SetOutput(stderr)
	file := fs.String("file", "", "path to rbx-sync.rbxmx")
	if err := fs.Parse(args[1:]); err != nil {
		return 1
	}
	src, err := resolveSource(env, *file)
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 1
	}
	dest, err := resolveDest(env)
	if err != nil {
		fmt.Fprintln(stderr, expectedWindowsDir)
		return 1
	}
	if err := copyFile(src, dest); err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 1
	}
	return 0
}

func resolveSource(env Env, fileFlag string) (string, error) {
	if fileFlag != "" {
		if _, err := os.Stat(fileFlag); err != nil {
			return "", err
		}
		return fileFlag, nil
	}
	beside := filepath.Join(env.ExecutableDir, pluginFileName)
	if fileExists(beside) {
		return beside, nil
	}
	if env.GOOS != "windows" && env.Home != "" {
		share := filepath.Join(env.Home, ".local", "share", "rbx-sync", pluginFileName)
		if fileExists(share) {
			return share, nil
		}
	}
	return "", errors.New("rbx-sync.rbxmx not found")
}

func resolveDest(env Env) (string, error) {
	if env.GOOS == "windows" {
		if env.LocalAppData == "" {
			return "", errDestMissing
		}
		return filepath.Join(env.LocalAppData, "Roblox", "Plugins", pluginFileName), nil
	}
	if dest, ok := destFromWindowsUser(env); ok {
		return dest, nil
	}
	if env.Glob != nil {
		matches, err := env.Glob(mntPluginsGlob)
		if err == nil && len(matches) == 1 {
			return filepath.Join(matches[0], pluginFileName), nil
		}
	}
	return "", errDestMissing
}

func destFromWindowsUser(env Env) (string, bool) {
	if env.LookPath == nil || env.CombinedOutput == nil {
		return "", false
	}
	if _, err := env.LookPath("cmd.exe"); err != nil {
		return "", false
	}
	out, err := env.CombinedOutput("cmd.exe", "/c", "echo %LOCALAPPDATA%")
	if err != nil {
		return "", false
	}
	win := strings.TrimSpace(string(out))
	if win == "" || win == "%LOCALAPPDATA%" {
		return "", false
	}
	unixOut, err := env.CombinedOutput("wslpath", "-u", win)
	if err != nil {
		return "", false
	}
	base := strings.TrimSpace(string(unixOut))
	if base == "" {
		return "", false
	}
	return path.Join(base, "Roblox", "Plugins", pluginFileName), true
}

func copyFile(src, dest string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dest, data, 0o644)
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

var errDestMissing = errors.New("plugin dest missing")
