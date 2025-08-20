package main

import (
	"fmt"
	"log"
	
	"github.com/setupkit"
)

func main() {
	// Enterprise-branded installer example
	err := setupkit.Install(setupkit.Config{
		AppName:   "Enterprise App",
		Version:   "2.0.0",
		Publisher: "ACME Corporation",
		Website:   "https://acme.corp",
		License:   "Proprietary License\n\nThis software is property of ACME Corp.",

		// Custom theme with corporate branding
		Theme: setupkit.Theme{
			Name:         "corporate",
			PrimaryColor: "#FF6B00", // Corporate Orange
			CustomCSS: `
				.header {
					background: linear-gradient(135deg, #FF6B00 0%, #FF8C00 100%);
				}
				.step.active .step-circle {
					background: linear-gradient(135deg, #FF6B00 0%, #FF8C00 100%);
					border-color: #FF6B00;
				}
				.button-primary {
					background: linear-gradient(135deg, #FF6B00 0%, #FF8C00 100%);
				}
			`,
		},

		Components: []setupkit.Component{
			{
				ID:       "server",
				Name:     "Server Components",
				Size:     250 * setupkit.MB,
				Required: true,
				Selected: true,
			},
			{
				ID:          "client",
				Name:        "Client Tools",
				Description: "Desktop client applications and utilities",
				Size:        85 * setupkit.MB,
				Selected:    true,
			},
			{
				ID:          "sdk",
				Name:        "Development SDK",
				Description: "SDK for developing custom extensions",
				Size:        120 * setupkit.MB,
				Selected:    false,
			},
		},

		// Custom callbacks for enterprise logging
		OnProgress: func(percent int, status string) {
			fmt.Printf("[%d%%] %s\n", percent, status)
		},

		OnComplete: func(path string) {
			fmt.Printf("✓ Installation completed at: %s\n", path)
		},

		BeforeInstall: func() error {
			fmt.Println("Preparing enterprise installation...")
			return nil
		},

		AfterInstall: func() error {
			fmt.Println("Finalizing installation...")
			return nil
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
