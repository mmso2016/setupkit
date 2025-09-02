package main

import (
	"log"

	"github.com/mmso2016/setupkit/installer"
	"github.com/mmso2016/setupkit/installer/core"
	// Import UI factory for GUI support
	_ "github.com/mmso2016/setupkit/installer/ui"
	// Import Wails UI for GUI wizard
	_ "github.com/mmso2016/setupkit/installer/ui/wails"
)

func main() {
	// GUI installer example using DFA-based wizard
	app, err := installer.New(
		// Basic application information
		installer.WithAppName("SetupKit Demo"),
		installer.WithVersion("1.0.0"),
		installer.WithPublisher("SetupKit Team"),
		installer.WithWebsite("https://github.com/mmso2016/setupkit"),
		installer.WithLicense(`MIT License

Copyright (c) 2025 SetupKit Team

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.`),

		// Installation components
		installer.WithComponents(
			core.Component{
				ID:       "core",
				Name:     "Core Application",
				Size:     45 * 1024 * 1024, // 45 MB
				Required: true,
				Selected: true,
			},
			core.Component{
				ID:          "docs",
				Name:        "Documentation",
				Description: "User manual and API documentation",
				Size:        12 * 1024 * 1024, // 12 MB
				Selected:    true,
			},
			core.Component{
				ID:          "examples",
				Name:        "Example Projects", 
				Description: "Sample projects and templates",
				Size:        8 * 1024 * 1024, // 8 MB
				Selected:    false,
			},
		),

		// Enable GUI mode with DFA-based wizard
		// This replaces hardcoded JavaScript navigation with DFA state management
		installer.WithMode(installer.ModeGUI),
		installer.WithDFAWizard(),
		
		// Set install directory
		installer.WithInstallDir("C:\\Program Files\\SetupKit Demo"),
		
		// Enable dry run for demo
		installer.WithDryRun(true),
		
		// Enable verbose logging
		installer.WithVerbose(true),
	)

	if err != nil {
		log.Fatal(err)
	}

	// Show DFA wizard status
	if app.IsUsingDFAWizard() {
		log.Println("✅ DFA Wizard enabled - GUI uses DFA instead of hardcoded JavaScript!")
	} else {
		log.Println("❌ Using legacy wizard")
	}

	// Run the installer with DFA-based GUI wizard
	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}