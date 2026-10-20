package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"roblox-sync/sync-agent/internal/packageplugin"
)

func main() {
	src := flag.String("src", filepath.Join("..", "studio-plugin", "src"), "studio-plugin source directory")
	out := flag.String("out", "rbx-sync.rbxmx", "output rbxmx path")
	flag.Parse()
	if err := packageplugin.Package(*src, *out); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
