package main

import (
	"context"
	"fmt"
	"gopherwm/xwindowmgr"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot determine config directory: %v\n", err)
		os.Exit(1)
	}
	configPath := filepath.Join(configDir, "gopherwm", "config.json")
	cfg, err := xwindowmgr.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error at loading configuration: %q: %v\n", configPath, err)
		os.Exit(1)
	}

	xManager, err := xwindowmgr.NewXWindowManager(cfg)
	if err != nil {
		fmt.Println("Error at creation XWindowManager:", err)
		os.Exit(1)
	}

	xManager.Start(ctx)

	// wait till Done
	<-ctx.Done()
	xManager.Wait()
}
