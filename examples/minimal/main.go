package main

import (
	"log"

	"github.com/setupkit/pkg/installer"
)

func main() {
	// Simple installer example using the framework
	err := installer.Install(installer.Config{
		AppName:   "My Application",
		Version:   "1.0.0",
		Publisher: "Example Corp",
		Website:   "https://example.com",
		License: `MIT License

Copyright (c) 2025 Example Corp

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
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.`,
		Components: []installer.Component{
			{
				ID:       "core",
				Name:     "Core Application",
				Size:     45 * 1024 * 1024, // 45 MB
				Required: true,
				Selected: true,
			},
			{
				ID:          "docs",
				Name:        "Documentation",
				Description: "User manual and API documentation",
				Size:        12 * 1024 * 1024, // 12 MB
				Selected:    true,
			},
			{
				ID:          "examples",
				Name:        "Example Projects",
				Description: "Sample projects and templates",
				Size:        8 * 1024 * 1024, // 8 MB
				Selected:    false,
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
