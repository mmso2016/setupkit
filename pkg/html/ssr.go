// Package html - SSR integration for SetupKit installer views
package html

import (
	"fmt"
	"github.com/mmso2016/setupkit/pkg/installer/core"
)

// SSRRenderer provides server-side rendering using the HTML builder
type SSRRenderer struct {
	// Configuration for the renderer
	theme string
}

// NewSSRRenderer creates a new SSR renderer
func NewSSRRenderer() *SSRRenderer {
	return &SSRRenderer{
		theme: "default",
	}
}

// SetTheme sets the theme for rendering
func (r *SSRRenderer) SetTheme(theme string) {
	r.theme = theme
}

// RenderWelcomePage renders the welcome/start page
func (r *SSRRenderer) RenderWelcomePage(config *core.Config) *Document {
	doc := NewDocument().
		SetTitle(config.AppName + " Setup").
		SetCharset("utf-8").
		SetViewport("").
		AddDefaultSetupKitStyles()

	// Main container
	container := DIV().Class("container").Children(
		// Header section
		HEADER().Class("header").Children(
			DIV().Class("title").Text(config.AppName + " Setup"),
			DIV().Class("subtitle").Text("Version: " + config.Version),
			DIV().Class("version").Text("Publisher: " + config.Publisher),
		),

		// Welcome content
		MAIN().Children(
			P("Welcome to the " + config.AppName + " installation wizard."),
			P("This wizard will guide you through the installation process."),
			BR(),
			P("Click Next to continue or Cancel to exit the installer."),
		),

		// Button section
		DIV().Class("buttons").Style("text-align: center; margin-top: 40px;").Children(
			BUTTON("Cancel").Class("button").ID("btnCancel"),
			BUTTON("Next").Class("button primary").ID("btnNext"),
		),
	)

	doc.AddToBody(container)
	
	// Add JavaScript for button interactions
	js := `
		document.addEventListener('DOMContentLoaded', function() {
			const btnNext = document.getElementById('btnNext');
			const btnCancel = document.getElementById('btnCancel');
			
			if (btnNext) {
				btnNext.addEventListener('click', function() {
					fetch('/api/next', { method: 'POST' })
						.then(response => response.json())
						.then(data => {
							if (data.status === 'ok') {
								window.location.reload();
							}
						});
				});
			}
			
			if (btnCancel) {
				btnCancel.addEventListener('click', function() {
					if (confirm('Are you sure you want to cancel the installation?')) {
						fetch('/api/cancel', { method: 'POST' })
							.then(response => response.json())
							.then(data => {
								window.close();
							});
					}
				});
			}
		});
	`
	
	doc.AddJS(js)
	return doc
}

// RenderLicensePage renders the license agreement page
func (r *SSRRenderer) RenderLicensePage(config *core.Config, license string) *Document {
	doc := NewDocument().
		SetTitle(config.AppName + " - License Agreement").
		SetCharset("utf-8").
		SetViewport("").
		AddDefaultSetupKitStyles()

	// License text container
	licenseDiv := DIV().Class("license-text").Style("height: 300px; overflow-y: scroll; padding: 15px; background: rgba(255,255,255,0.1); border-radius: 10px; margin: 20px 0;").Child(
		PRE(license).Style("white-space: pre-wrap; font-family: monospace; color: #333;"),
	)

	// Acceptance checkbox
	checkboxDiv := DIV().Class("license-acceptance").Style("margin: 20px 0; text-align: center;").Children(
		INPUT("checkbox").ID("acceptLicense").Style("margin-right: 10px;"),
		LABEL("I accept the terms of the license agreement").Attr("for", "acceptLicense"),
	)

	// Main container
	container := DIV().Class("container").Children(
		// Header
		HEADER().Class("header").Children(
			DIV().Class("title").Text("License Agreement"),
			DIV().Class("subtitle").Text("Please read and accept the license"),
		),

		// License content
		licenseDiv,

		// Acceptance
		checkboxDiv,

		// Buttons
		DIV().Class("buttons").Style("text-align: center; margin-top: 40px;").Children(
			BUTTON("Back").Class("button").ID("btnBack"),
			BUTTON("Next").Class("button primary").ID("btnNext").Attr("disabled", "true"),
			BUTTON("Cancel").Class("button").ID("btnCancel"),
		),
	)

	doc.AddToBody(container)

	// Add JavaScript for license acceptance and navigation
	js := `
		document.addEventListener('DOMContentLoaded', function() {
			const acceptCheckbox = document.getElementById('acceptLicense');
			const btnNext = document.getElementById('btnNext');
			const btnBack = document.getElementById('btnBack');
			const btnCancel = document.getElementById('btnCancel');

			// Enable/disable Next button based on license acceptance
			if (acceptCheckbox && btnNext) {
				acceptCheckbox.addEventListener('change', function() {
					if (this.checked) {
						btnNext.removeAttribute('disabled');
						btnNext.classList.add('primary');
					} else {
						btnNext.setAttribute('disabled', 'true');
						btnNext.classList.remove('primary');
					}
				});
			}

			if (btnNext) {
				btnNext.addEventListener('click', function() {
					const accepted = acceptCheckbox && acceptCheckbox.checked;
					if (accepted) {
						fetch('/api/license', {
							method: 'POST',
							headers: {'Content-Type': 'application/x-www-form-urlencoded'},
							body: 'accepted=true'
						})
						.then(response => response.json())
						.then(data => {
							if (data.status === 'ok') {
								window.location.reload();
							}
						});
					}
				});
			}

			if (btnBack) {
				btnBack.addEventListener('click', function() {
					fetch('/api/prev', { method: 'POST' })
						.then(response => response.json())
						.then(data => {
							if (data.status === 'ok') {
								window.location.reload();
							}
						});
				});
			}

			if (btnCancel) {
				btnCancel.addEventListener('click', function() {
					if (confirm('Are you sure you want to cancel the installation?')) {
						fetch('/api/cancel', { method: 'POST' })
							.then(response => response.json())
							.then(data => {
								window.close();
							});
					}
				});
			}
		});
	`

	doc.AddJS(js)
	return doc
}

// RenderInstallPathPage renders the installation path selection page
func (r *SSRRenderer) RenderInstallPathPage(config *core.Config, defaultPath string) *Document {
	doc := NewDocument().
		SetTitle(config.AppName + " - Installation Path").
		SetCharset("utf-8").
		SetViewport("").
		AddDefaultSetupKitStyles()

	// Path selection form
	pathDiv := DIV().Class("path-selection").Style("margin: 30px 0;").Children(
		DIV().Class("form-group").Children(
			LABEL("Installation directory:").Attr("for", "installPath").Style("display: block; margin-bottom: 10px; font-weight: bold;"),
			INPUT("text").ID("installPath").Attr("value", defaultPath).Style("width: 100%; padding: 10px; border: 1px solid #ccc; border-radius: 5px; font-size: 1rem;"),
		),
		DIV().Class("browse-button").Style("margin-top: 10px;").Child(
			BUTTON("Browse...").Class("button").ID("btnBrowse").Style("padding: 8px 16px;"),
		),
	)

	// Space info (simplified - would be dynamic in real implementation)
	infoDiv := DIV().Class("path-info").Style("margin: 20px 0; padding: 15px; background: rgba(255,255,255,0.1); border-radius: 10px;").Children(
		H4("Installation Information"),
		P("Required space: 50 MB"),
		P("Available space: 2.5 GB"),
	)

	// Main container
	container := DIV().Class("container").Children(
		// Header
		HEADER().Class("header").Children(
			DIV().Class("title").Text("Installation Directory"),
			DIV().Class("subtitle").Text("Choose where to install " + config.AppName),
		),

		// Path selection
		pathDiv,

		// Info
		infoDiv,

		// Buttons
		DIV().Class("buttons").Style("text-align: center; margin-top: 40px;").Children(
			BUTTON("Back").Class("button").ID("btnBack"),
			BUTTON("Next").Class("button primary").ID("btnNext"),
			BUTTON("Cancel").Class("button").ID("btnCancel"),
		),
	)

	doc.AddToBody(container)

	// Add JavaScript for path selection and navigation
	js := `
		document.addEventListener('DOMContentLoaded', function() {
			const installPathInput = document.getElementById('installPath');
			const btnBrowse = document.getElementById('btnBrowse');
			const btnNext = document.getElementById('btnNext');
			const btnBack = document.getElementById('btnBack');
			const btnCancel = document.getElementById('btnCancel');

			// Browse button (simplified - would use file dialog in real implementation)
			if (btnBrowse) {
				btnBrowse.addEventListener('click', function() {
					alert('File browser not implemented in demo. Please type the path manually.');
				});
			}

			if (btnNext) {
				btnNext.addEventListener('click', function() {
					const path = installPathInput.value.trim();
					if (path) {
						fetch('/api/path', {
							method: 'POST',
							headers: {'Content-Type': 'application/x-www-form-urlencoded'},
							body: 'path=' + encodeURIComponent(path)
						})
						.then(response => response.json())
						.then(data => {
							if (data.status === 'ok') {
								window.location.reload();
							}
						});
					} else {
						alert('Please enter an installation path.');
					}
				});
			}

			if (btnBack) {
				btnBack.addEventListener('click', function() {
					fetch('/api/prev', { method: 'POST' })
						.then(response => response.json())
						.then(data => {
							if (data.status === 'ok') {
								window.location.reload();
							}
						});
				});
			}

			if (btnCancel) {
				btnCancel.addEventListener('click', function() {
					if (confirm('Are you sure you want to cancel the installation?')) {
						fetch('/api/cancel', { method: 'POST' })
							.then(response => response.json())
							.then(data => {
								window.close();
							});
					}
				});
			}
		});
	`

	doc.AddJS(js)
	return doc
}

// RenderSummaryPage renders the installation summary page
func (r *SSRRenderer) RenderSummaryPage(config *core.Config, selectedComponents []core.Component, installPath string) *Document {
	doc := NewDocument().
		SetTitle(config.AppName + " - Installation Summary").
		SetCharset("utf-8").
		SetViewport("").
		AddDefaultSetupKitStyles()

	// Calculate totals
	var totalSize int64
	for _, comp := range selectedComponents {
		totalSize += comp.Size
	}

	// Summary sections
	appInfoDiv := DIV().Class("summary-section").Style("margin: 20px 0; padding: 20px; background: rgba(255,255,255,0.1); border-radius: 10px;").Children(
		H3("Application Information"),
		DIV().Style("display: grid; grid-template-columns: 150px 1fr; gap: 10px;").Children(
			SPAN("Name:").Style("font-weight: bold;"),
			SPAN(config.AppName),
			SPAN("Version:").Style("font-weight: bold;"),
			SPAN(config.Version),
			SPAN("Publisher:").Style("font-weight: bold;"),
			SPAN(config.Publisher),
			SPAN("Install Path:").Style("font-weight: bold;"),
			SPAN(installPath),
		),
	)

	// Components list
	componentsDiv := DIV().Class("summary-section").Style("margin: 20px 0; padding: 20px; background: rgba(255,255,255,0.1); border-radius: 10px;").Children(
		H3(fmt.Sprintf("Selected Components (%d)", len(selectedComponents))),
	)

	componentsList := UL().Style("list-style: none; padding: 0;")
	for _, comp := range selectedComponents {
		marker := "•"
		if comp.Required {
			marker = "🔒"
		}

		listItem := LI().Style("margin: 5px 0; display: flex; justify-content: space-between;").Children(
			SPAN(marker + " " + comp.Name),
			SPAN(formatSize(comp.Size)).Style("opacity: 0.8;"),
		)
		componentsList.Child(listItem)
	}
	componentsDiv.Child(componentsList)

	// Space requirements
	spaceDiv := DIV().Class("summary-section").Style("margin: 20px 0; padding: 20px; background: rgba(255,255,255,0.1); border-radius: 10px;").Children(
		H3("Disk Space Requirements"),
		DIV().Style("display: grid; grid-template-columns: 200px 1fr; gap: 10px;").Children(
			SPAN("Total size required:").Style("font-weight: bold;"),
			SPAN(formatSize(totalSize)),
			SPAN("Available space:").Style("font-weight: bold;"),
			SPAN("2.5 GB").Style("color: #4CAF50;"), // Simplified - would be dynamic
		),
	)

	// Main container
	container := DIV().Class("container").Children(
		// Header
		HEADER().Class("header").Children(
			DIV().Class("title").Text("Ready to Install"),
			DIV().Class("subtitle").Text("Review your installation settings"),
		),

		// Summary sections
		appInfoDiv,
		componentsDiv,
		spaceDiv,

		// Warning/confirmation
		DIV().Class("confirmation").Style("margin: 30px 0; text-align: center; padding: 15px; background: rgba(255, 193, 7, 0.2); border-radius: 10px;").Child(
			P("Click Install to begin the installation process.").Style("margin: 0; font-weight: bold;"),
		),

		// Buttons
		DIV().Class("buttons").Style("text-align: center; margin-top: 40px;").Children(
			BUTTON("Back").Class("button").ID("btnBack"),
			BUTTON("Install").Class("button primary").ID("btnInstall").Style("font-size: 1.2rem; padding: 12px 30px;"),
			BUTTON("Cancel").Class("button").ID("btnCancel"),
		),
	)

	doc.AddToBody(container)

	// Add JavaScript for summary actions and navigation
	js := `
		document.addEventListener('DOMContentLoaded', function() {
			const btnInstall = document.getElementById('btnInstall');
			const btnBack = document.getElementById('btnBack');
			const btnCancel = document.getElementById('btnCancel');

			if (btnInstall) {
				btnInstall.addEventListener('click', function() {
					if (confirm('Begin installation now?')) {
						btnInstall.textContent = 'Installing...';
						btnInstall.disabled = true;

						fetch('/api/next', { method: 'POST' })
							.then(response => response.json())
							.then(data => {
								if (data.status === 'ok') {
									window.location.reload();
								}
							});
					}
				});
			}

			if (btnBack) {
				btnBack.addEventListener('click', function() {
					fetch('/api/prev', { method: 'POST' })
						.then(response => response.json())
						.then(data => {
							if (data.status === 'ok') {
								window.location.reload();
							}
						});
				});
			}

			if (btnCancel) {
				btnCancel.addEventListener('click', function() {
					if (confirm('Are you sure you want to cancel the installation?')) {
						fetch('/api/cancel', { method: 'POST' })
							.then(response => response.json())
							.then(data => {
								window.close();
							});
					}
				});
			}
		});
	`

	doc.AddJS(js)
	return doc
}

// RenderComponentsPage renders the component selection page
func (r *SSRRenderer) RenderComponentsPage(config *core.Config) *Document {
	doc := NewDocument().
		SetTitle(config.AppName + " - Component Selection").
		SetCharset("utf-8").
		SetViewport("").
		AddDefaultSetupKitStyles()

	// Build components list
	componentsDiv := DIV().Class("components")
	
	var totalSize int64
	selectedCount := 0
	
	for i, comp := range config.Components {
		totalSize += comp.Size
		if comp.Selected {
			selectedCount++
		}
		
		// Component container
		compDiv := DIV().Class("component")
		if comp.Required {
			compDiv.Class("component required")
		}
		
		// Component header with checkbox simulation
		checkmark := "☐"
		if comp.Selected {
			checkmark = "☑"
		}
		if comp.Required {
			checkmark = "🔒"
		}
		
		compHeader := DIV().Class("component-header").Children(
			SPAN(checkmark).Style("margin-right: 15px; font-size: 1.2rem;"),
			DIV().Class("component-name").Text(comp.Name),
			DIV().Class("component-size").Text(formatSize(comp.Size)),
		)
		
		compDiv.Child(compHeader)
		
		// Component description
		if comp.Description != "" {
			compDiv.Child(
				DIV().Class("component-description").Text(comp.Description),
			)
		}
		
		// Add data attributes for JavaScript interaction
		compDiv.Attr("data-component-id", comp.ID).
			Attr("data-component-index", fmt.Sprintf("%d", i)).
			Attr("data-selected", boolToString(comp.Selected)).
			Attr("data-required", boolToString(comp.Required))
		
		componentsDiv.Child(compDiv)
	}

	// Summary section
	summaryDiv := DIV().Class("summary").Style("margin-top: 30px; padding: 20px; background: rgba(255,255,255,0.1); border-radius: 10px;").Children(
		H3("Installation Summary"),
		P(fmt.Sprintf("Selected components: %d of %d", selectedCount, len(config.Components))),
		P("Total size: " + formatSize(totalSize)),
	)

	// Main container
	container := DIV().Class("container").Children(
		// Header
		HEADER().Class("header").Children(
			DIV().Class("title").Text("Component Selection"),
			DIV().Class("subtitle").Text("Choose components to install"),
		),

		// Components list
		componentsDiv,
		
		// Summary
		summaryDiv,

		// Instructions
		DIV().Class("instructions").Style("margin-top: 20px; opacity: 0.8;").Children(
			P("Click on components to select or deselect them."),
			P("Required components cannot be deselected."),
		),

		// Buttons
		DIV().Class("buttons").Style("text-align: center; margin-top: 40px;").Children(
			BUTTON("Back").Class("button").ID("btnBack"),
			BUTTON("Next").Class("button primary").ID("btnNext"),
			BUTTON("Cancel").Class("button").ID("btnCancel"),
		),
	)

	doc.AddToBody(container)
	
	// Add JavaScript for component interaction and navigation
	js := `
		document.addEventListener('DOMContentLoaded', function() {
			// Component selection logic
			const components = document.querySelectorAll('.component:not(.required)');
			components.forEach(comp => {
				comp.addEventListener('click', function() {
					const isSelected = this.getAttribute('data-selected') === 'true';
					const newSelected = !isSelected;
					
					this.setAttribute('data-selected', newSelected.toString());
					
					// Update checkmark
					const checkmark = this.querySelector('span');
					checkmark.textContent = newSelected ? '☑' : '☐';
					
					// Update summary (simplified)
					updateSummary();
				});
				
				comp.style.cursor = 'pointer';
			});
			
			// Button navigation logic
			const btnNext = document.getElementById('btnNext');
			const btnBack = document.getElementById('btnBack');
			const btnCancel = document.getElementById('btnCancel');
			
			if (btnNext) {
				btnNext.addEventListener('click', function() {
					fetch('/api/next', { method: 'POST' })
						.then(response => response.json())
						.then(data => {
							if (data.status === 'ok') {
								window.location.reload();
							}
						});
				});
			}
			
			if (btnBack) {
				btnBack.addEventListener('click', function() {
					fetch('/api/prev', { method: 'POST' })
						.then(response => response.json())
						.then(data => {
							if (data.status === 'ok') {
								window.location.reload();
							}
						});
				});
			}
			
			if (btnCancel) {
				btnCancel.addEventListener('click', function() {
					if (confirm('Are you sure you want to cancel the installation?')) {
						fetch('/api/cancel', { method: 'POST' })
							.then(response => response.json())
							.then(data => {
								window.close();
							});
					}
				});
			}
		});
		
		function updateSummary() {
			// This would update the summary section with new totals
			// Implementation depends on the specific needs
		}
	`
	
	doc.AddJS(js)
	return doc
}

// RenderProgressPage renders the installation progress page
func (r *SSRRenderer) RenderProgressPage(config *core.Config, progress int, status string) *Document {
	doc := NewDocument().
		SetTitle(config.AppName + " - Installing").
		SetCharset("utf-8").
		SetViewport("").
		AddDefaultSetupKitStyles()

	// Progress bar
	progressDiv := DIV().Class("progress").Child(
		DIV().Class("progress-bar").Style(fmt.Sprintf("width: %d%%;", progress)),
	)

	// Status text
	statusDiv := DIV().Class("status").Style("text-align: center; margin: 20px 0;").Children(
		H3("Installing " + config.AppName),
		P(status),
		P(fmt.Sprintf("%d%% complete", progress)),
	)

	container := DIV().Class("container").Children(
		HEADER().Class("header").Children(
			DIV().Class("title").Text("Installation in Progress"),
		),
		statusDiv,
		progressDiv,
		DIV().Style("text-align: center; margin-top: 40px;").Child(
			P("Please wait while the installation completes..."),
		),
	)

	doc.AddToBody(container)
	return doc
}

// RenderCompletionPage renders the installation completion page
func (r *SSRRenderer) RenderCompletionPage(config *core.Config, success bool) *Document {
	title := "Installation Complete"
	if !success {
		title = "Installation Failed"
	}

	doc := NewDocument().
		SetTitle(config.AppName + " - " + title).
		SetCharset("utf-8").
		SetViewport("").
		AddDefaultSetupKitStyles()

	var message, icon string
	if success {
		message = config.AppName + " has been successfully installed."
		icon = "✅"
	} else {
		message = "The installation of " + config.AppName + " was not completed successfully."
		icon = "❌"
	}

	container := DIV().Class("container").Children(
		HEADER().Class("header").Children(
			DIV().Style("font-size: 4rem; margin-bottom: 20px;").Text(icon),
			DIV().Class("title").Text(title),
		),
		MAIN().Style("text-align: center;").Children(
			P(message).Style("font-size: 1.2rem; margin-bottom: 30px;"),
		),
		DIV().Class("buttons").Style("text-align: center;").Child(
			BUTTON("Finish").Class("button primary").ID("btnFinish"),
		),
	)

	doc.AddToBody(container)
	
	// Add JavaScript for finish button
	js := `
		document.addEventListener('DOMContentLoaded', function() {
			const btnFinish = document.getElementById('btnFinish');
			
			if (btnFinish) {
				btnFinish.addEventListener('click', function() {
					fetch('/api/finish', { method: 'POST' })
						.then(response => response.json())
						.then(data => {
							window.close();
						});
				});
			}
		});
	`
	
	doc.AddJS(js)
	return doc
}

// Helper functions

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}