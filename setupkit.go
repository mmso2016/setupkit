// Package setupkit provides a modern, easy-to-use installer framework for Windows applications.
// This is the main API entry point that provides a simple interface to the underlying installer framework.
package setupkit

import (
	"github.com/mmso2016/setupkit/pkg/installer"
	"github.com/mmso2016/setupkit/pkg/installer/core"
)

// Config is an alias for core.Config
type Config = core.Config

// Component is an alias for core.Component
type Component = core.Component

// InstallMode is an alias for core.InstallMode
type InstallMode = core.InstallMode

// Mode constants
const (
	ModeExpress     = core.ModeExpress
	ModeCustom      = core.ModeCustom
	ModeAdvanced    = core.ModeAdvanced
	ModeRepair      = core.ModeRepair
	ModeUninstall   = core.ModeUninstall
	ModeUserDefined = core.ModeUserDefined
)

// UI Mode constants
const (
	ModeSilent = core.ModeSilent
	ModeCLI    = core.ModeCLI
	ModeGUI    = core.ModeGUI
	ModeAuto   = core.ModeAuto
)

// Install creates and runs an installer with the given configuration.
// This is the simplest way to create an installer - just call this function.
//
// Example:
//
//	setupkit.Install(&setupkit.Config{
//	    AppName: "My App",
//	    Version: "1.0.0",
//	    Components: []setupkit.Component{
//	        {ID: "core", Name: "Core Files", Required: true},
//	    },
//	})
func Install(config *Config) error {
	app, err := New(config)
	if err != nil {
		return err
	}
	return app.Run()
}

// New creates a new installer instance with the given configuration.
// This gives you more control over the installer setup process.
//
// Example:
//
//	app, err := setupkit.New(&setupkit.Config{
//	    AppName: "My App",
//	    Version: "1.0.0",
//	})
//	if err != nil {
//	    return err
//	}
//	return app.Run()
func New(config *Config) (*installer.Installer, error) {
	// Convert setupkit config to installer options
	opts := []installer.Option{}
	
	if config.AppName != "" {
		opts = append(opts, installer.WithAppName(config.AppName))
	}
	if config.Version != "" {
		opts = append(opts, installer.WithVersion(config.Version))
	}
	if config.Publisher != "" {
		opts = append(opts, installer.WithPublisher(config.Publisher))
	}
	if config.Website != "" {
		opts = append(opts, installer.WithWebsite(config.Website))
	}
	if config.InstallDir != "" {
		opts = append(opts, installer.WithInstallDir(config.InstallDir))
	}
	if len(config.Components) > 0 {
		opts = append(opts, installer.WithComponents(config.Components...))
	}
	if config.Mode != 0 {
		opts = append(opts, installer.WithMode(config.Mode))
	}
	
	return installer.New(opts...)
}

// Common size constants for convenience
const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
)

// Common licenses for convenience
const (
	LicenseMIT = `MIT License

Copyright (c) 2025

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
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.`

	LicenseApache2 = `Apache License, Version 2.0

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.`

	LicenseGPL3 = `GNU General Public License v3.0

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.`
)