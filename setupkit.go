// Package setupkit provides a modern, easy-to-use installer framework for Windows applications.
package setupkit

import (
	"github.com/mmso2016/setupkit/pkg/installer"
)

// Config is an alias for installer.Config
type Config = installer.Config

// Component is an alias for installer.Component
type Component = installer.Component

// Theme is an alias for installer.Theme
type Theme = installer.Theme

// Install creates and runs an installer with the given configuration.
// This is the simplest way to create an installer - just call this function.
//
// Example:
//
//	setupkit.Install(setupkit.Config{
//	    AppName: "My App",
//	    Version: "1.0.0",
//	    Components: []setupkit.Component{
//	        {ID: "core", Name: "Core Files", Required: true},
//	    },
//	})
func Install(config Config) error {
	return installer.Install(config)
}

// Run creates and runs an installer with the given configuration.
// This is the recommended way to use SetupKit.
//
// Example:
//
//	setupkit.Run(&setupkit.Config{
//	    AppName: "My App",
//	    Version: "1.0.0",
//	})
func Run(config *Config) error {
	return installer.Run(config)
}

// Common size constants for convenience
const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
)

// Predefined themes
var (
	ThemeDefault = Theme{Name: "default"}
	ThemeDark    = Theme{
		Name: "dark",
		CustomCSS: `
			body { background: #1a1a2e; }
			.installer-container { background: #16213e; color: #eee; }
			.header { background: linear-gradient(135deg, #0f3460 0%, #16213e 100%); }
		`,
	}
	ThemeCorporate = Theme{
		Name:         "corporate",
		PrimaryColor: "#0061a7",
		CustomCSS: `
			.header { background: linear-gradient(135deg, #0061a7 0%, #004d84 100%); }
			.step.active .step-circle { background: #0061a7; }
		`,
	}
)

// Common licenses
const (
	LicenseMIT = `MIT License

Copyright (c) 2024

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
