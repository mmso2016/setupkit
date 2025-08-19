package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/mmso2016/setupkit/installer/core"
	_ "github.com/mmso2016/setupkit/installer/ui" // Register UI factory
)

func main() {
	fmt.Println("SetupKit GUI Alternative (Console Mode)")
	fmt.Println("============================================")
	fmt.Println()

	// Configuration
	config := &core.Config{
		AppName:    "Go Installer Example",
		Version:    "1.0.0",
		Publisher:  "Go Installer Team",
		InstallDir: filepath.Join("C:", "Program Files", "SetupKit-Sample"),
		Mode:       core.ModeCLI,
		Components: []core.Component{
			{
				ID:          "core",
				Name:        "Core Files",
				Description: "Required application files",
				Required:    true,
				Selected:    true,
				Size:        10485760, // 10 MB
				Installer: func(ctx context.Context) error {
					fmt.Println("Installing core files...")
					time.Sleep(time.Second)
					return nil
				},
			},
			{
				ID:          "docs",
				Name:        "Documentation",
				Description: "User manual and API documentation",
				Required:    false,
				Selected:    true,
				Size:        5242880, // 5 MB
				Installer: func(ctx context.Context) error {
					fmt.Println("Installing documentation...")
					time.Sleep(time.Second)
					return nil
				},
			},
		},
		License: `MIT License

Copyright (c) 2025 Go SetupKit

Permission is hereby granted...`,
	}

	// Create and run installer
	core.New(config)
	ctx := context.Background()

	// Simulate installation flow
	fmt.Println("1. Welcome to Go SetupKit")
	fmt.Println("2. License Agreement: [Accepted]")
	fmt.Println("3. Components:")
	for _, comp := range config.Components {
		status := "[ ]"
		if comp.Selected {
			status = "[X]"
		}
		fmt.Printf("   %s %s - %s\n", status, comp.Name, comp.Description)
	}
	fmt.Printf("4. Installation Path: %s\n", config.InstallDir)
	fmt.Println("5. Installing...")

	// Run component installers
	for _, comp := range config.Components {
		if comp.Selected && comp.Installer != nil {
			if err := comp.Installer(ctx); err != nil {
				log.Fatalf("Failed to install %s: %v", comp.Name, err)
			}
		}
	}

	fmt.Println("\nInstallation completed successfully!")
}
