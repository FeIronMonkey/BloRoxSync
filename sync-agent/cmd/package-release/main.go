package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"roblox-sync/sync-agent/internal/release"
)

func main() {
	out := flag.String("out", "dist", "directory for release assets")
	flag.Parse()
	agentDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	pluginSrc := filepath.Join(agentDir, "..", "studio-plugin", "src")
	if err := release.Package(agentDir, pluginSrc, *out); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
