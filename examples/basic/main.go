package main

import (
	"context"
	"embed"
	"log"

	"github.com/mmso2016/setupkit/installer"
)

//go:embed assets/*
var assets embed.FS

func main() {
	// Create installer with options
	inst, err := installer.New(
		installer.WithAppName("MyApp"),
		installer.WithVersion("1.0.0"),
		installer.WithMode(installer.ModeAuto), // Auto-detect best UI
		installer.WithAssets(assets),
		installer.WithLicense("MIT License - see LICENSE file for details"),
		installer.WithComponents(
			installer.Component{
				ID:          "core",
				Name:        "Core Files",
				Description: "Essential application files",
				Required:    true,
				Selected:    true,
				Size:        10 * 1024 * 1024, // 10MB
				Installer:   installCore,
			},
			installer.Component{
				ID:          "docs",
				Name:        "Documentation",
				Description: "User manual and help files",
				Required:    false,
				Selected:    true,
				Size:        5 * 1024 * 1024, // 5MB
				Installer:   installDocs,
			},
		),
		installer.WithInstallDir("C:\\Program Files\\MyApp"),
		installer.WithVerbose(true),
	)

	if err != nil {
		log.Fatal("Failed to create installer:", err)
	}

	// Run the installer
	if err := inst.Run(); err != nil {
		log.Fatal("Installation failed:", err)
	}

	log.Println("Installation completed successfully!")
}

func installCore(ctx context.Context) error {
	// Get logger from context
	if logger, ok := ctx.Value("logger").(installer.Logger); ok {
		logger.Info("Installing core files...")
	}
	// TODO: Implement actual installation
	// Copy files from assets to install directory
	return nil
}

func installDocs(ctx context.Context) error {
	// Get logger from context
	if logger, ok := ctx.Value("logger").(installer.Logger); ok {
		logger.Info("Installing documentation...")
	}
	// TODO: Implement actual installation
	// Copy documentation files
	return nil
}
