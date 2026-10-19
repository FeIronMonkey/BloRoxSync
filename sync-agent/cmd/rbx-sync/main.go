package main

import (
	"context"
	"os"
	"os/signal"

	"roblox-sync/sync-agent/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	os.Exit(cli.Run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}
