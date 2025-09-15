package views

import (
	"html/template"
	texttemplate "text/template"
)

// Embedded HTML templates for WebView
var htmlTemplates = map[string]string{
	"welcome": `<!DOCTYPE html>
<html>
<head>
    <title>{{.AppName}} Setup</title>
    <style>
        body { font-family: 'Segoe UI', Arial, sans-serif; margin: 0; padding: 20px; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; }
        .container { max-width: 600px; margin: 0 auto; background: rgba(255,255,255,0.1); border-radius: 10px; padding: 40px; backdrop-filter: blur(10px); }
        .header { text-align: center; margin-bottom: 40px; }
        .logo { font-size: 48px; margin-bottom: 20px; }
        .title { font-size: 32px; font-weight: 300; margin-bottom: 10px; }
        .subtitle { font-size: 16px; opacity: 0.8; }
        .info { margin: 20px 0; }
        .info-item { margin: 10px 0; font-size: 16px; }
        .nav { display: flex; justify-content: space-between; margin-top: 40px; }
        .btn { padding: 12px 24px; border: none; border-radius: 6px; font-size: 16px; cursor: pointer; transition: all 0.3s; }
        .btn-primary { background: #4CAF50; color: white; }
        .btn-secondary { background: rgba(255,255,255,0.2); color: white; }
        .btn:hover { transform: translateY(-2px); box-shadow: 0 4px 8px rgba(0,0,0,0.2); }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">🚀</div>
            <h1 class="title">Welcome to {{.AppName}} Setup</h1>
            <div class="subtitle">Version {{.Version}}</div>
        </div>
        <div class="info">
            {{if .Publisher}}<div class="info-item">Publisher: {{.Publisher}}</div>{{end}}
            {{if .Website}}<div class="info-item">Website: <a href="{{.Website}}" style="color: #87CEEB;">{{.Website}}</a></div>{{end}}
            <div class="info-item">This wizard will guide you through the installation process.</div>
        </div>
        <div class="nav">
            <button class="btn btn-secondary" onclick="cancel()">{{.CancelLabel}}</button>
            <button class="btn btn-primary" onclick="next()">{{.NextLabel}}</button>
        </div>
    </div>
</body>
</html>`,

	"license": `<!DOCTYPE html>
<html>
<head>
    <title>{{.AppName}} - License Agreement</title>
    <style>
        body { font-family: 'Segoe UI', Arial, sans-serif; margin: 0; padding: 20px; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; }
        .container { max-width: 700px; margin: 0 auto; background: rgba(255,255,255,0.1); border-radius: 10px; padding: 30px; backdrop-filter: blur(10px); }
        .header { text-align: center; margin-bottom: 30px; }
        .license-box { background: rgba(255,255,255,0.9); color: #333; padding: 20px; border-radius: 8px; height: 300px; overflow-y: auto; margin: 20px 0; font-family: monospace; font-size: 14px; line-height: 1.6; }
        .checkbox-container { display: flex; align-items: center; margin: 20px 0; font-size: 16px; }
        .checkbox-container input { margin-right: 10px; transform: scale(1.2); }
        .nav { display: flex; justify-content: space-between; margin-top: 30px; }
        .btn { padding: 12px 24px; border: none; border-radius: 6px; font-size: 16px; cursor: pointer; transition: all 0.3s; }
        .btn-primary { background: #4CAF50; color: white; }
        .btn-secondary { background: rgba(255,255,255,0.2); color: white; }
        .btn:disabled { opacity: 0.5; cursor: not-allowed; }
        .btn:hover:not(:disabled) { transform: translateY(-2px); box-shadow: 0 4px 8px rgba(0,0,0,0.2); }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>License Agreement</h1>
            <p>Please review and accept the license agreement to continue</p>
        </div>
        <div class="license-box">{{.License}}</div>
        <div class="checkbox-container">
            <input type="checkbox" id="accept" onchange="toggleNext()">
            <label for="accept">I accept the terms of the license agreement</label>
        </div>
        <div class="nav">
            <button class="btn btn-secondary" onclick="back()">{{.BackLabel}}</button>
            <button class="btn btn-primary" id="nextBtn" disabled onclick="next()">{{.NextLabel}}</button>
        </div>
    </div>
    <script>
        function toggleNext() {
            document.getElementById('nextBtn').disabled = !document.getElementById('accept').checked;
        }
    </script>
</body>
</html>`,

	"components": `<!DOCTYPE html>
<html>
<head>
    <title>{{.AppName}} - Component Selection</title>
    <style>
        body { font-family: 'Segoe UI', Arial, sans-serif; margin: 0; padding: 20px; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; }
        .container { max-width: 700px; margin: 0 auto; background: rgba(255,255,255,0.1); border-radius: 10px; padding: 30px; backdrop-filter: blur(10px); }
        .header { text-align: center; margin-bottom: 30px; }
        .component { background: rgba(255,255,255,0.1); margin: 10px 0; padding: 20px; border-radius: 8px; border-left: 4px solid #4CAF50; }
        .component.required { border-left-color: #ff9800; }
        .component-header { display: flex; align-items: center; margin-bottom: 10px; }
        .component-header input { margin-right: 15px; transform: scale(1.3); }
        .component-name { font-size: 18px; font-weight: 500; }
        .component-size { margin-left: auto; font-size: 14px; opacity: 0.8; }
        .component-desc { font-size: 14px; opacity: 0.9; line-height: 1.4; }
        .summary { background: rgba(255,255,255,0.1); padding: 20px; border-radius: 8px; margin: 30px 0; }
        .nav { display: flex; justify-content: space-between; margin-top: 30px; }
        .btn { padding: 12px 24px; border: none; border-radius: 6px; font-size: 16px; cursor: pointer; transition: all 0.3s; }
        .btn-primary { background: #4CAF50; color: white; }
        .btn-secondary { background: rgba(255,255,255,0.2); color: white; }
        .btn:hover { transform: translateY(-2px); box-shadow: 0 4px 8px rgba(0,0,0,0.2); }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Component Selection</h1>
            <p>Choose which components to install</p>
        </div>
        
        {{range .Components}}
        <div class="component {{if .Required}}required{{end}}">
            <div class="component-header">
                <input type="checkbox" {{if .Selected}}checked{{end}} {{if .Required}}disabled{{end}} onchange="updateSummary()">
                <span class="component-name">{{.Name}}</span>
                <span class="component-size">{{.Size}}</span>
            </div>
            {{if .Description}}<div class="component-desc">{{.Description}}</div>{{end}}
        </div>
        {{end}}
        
        <div class="summary">
            <strong>Selected Components: <span id="selectedCount">{{len .SelectedComponents}}</span></strong><br>
            <strong>Total Size: <span id="totalSize">{{.TotalSize}}</span></strong>
        </div>
        
        <div class="nav">
            <button class="btn btn-secondary" onclick="back()">{{.BackLabel}}</button>
            <button class="btn btn-primary" onclick="next()">{{.NextLabel}}</button>
        </div>
    </div>
    <script>
        function updateSummary() {
            // Update component summary dynamically
        }
    </script>
</body>
</html>`,
}

// Embedded CLI templates
var cliTemplates = map[string]string{
	"welcome": `
{{separator 60}}
{{center .PageTitle 60}}
{{separator 60}}

  Welcome to {{.AppName}} Setup
  Version: {{.Version}}
{{if .Publisher}}  Publisher: {{.Publisher}}{{end}}
{{if .Website}}  Website: {{.Website}}{{end}}

  This wizard will guide you through the installation process.

{{separator 60}}`,

	"license": `
{{separator 60}}
{{center "License Agreement" 60}}
{{separator 60}}

{{.License}}

{{separator 60}}
Do you accept the license agreement? (y/n): `,

	"components": `
{{separator 60}}
{{center "Component Selection" 60}}
{{separator 60}}

Choose components to install:

{{range .Components}}
  [{{if .Required}}R{{else}}{{selected .Selected}}{{end}}] {{.Index}}. {{.Name}} ({{.Size}})
{{if .Description}}      {{.Description}}{{end}}
{{end}}

  R = Required, X = Selected

Total selected: {{len .SelectedComponents}} components ({{.TotalSize}})

Enter component numbers to toggle (comma-separated), or press Enter to continue:
`,

	"location": `
{{separator 60}}
{{center "Installation Location" 60}}
{{separator 60}}

Install location [{{.InstallPath}}]: `,

	"summary": `
{{separator 60}}
{{center "Installation Summary" 60}}
{{separator 60}}

Application: {{.AppName}} v{{.Version}}
Install to: {{.InstallPath}}
Components: {{len .SelectedComponents}} selected
Total size: {{.TotalSize}}

Selected components:
{{range .SelectedComponents}}  - {{.Name}} ({{.Size}})
{{end}}

{{separator 60}}
Proceed with installation? (y/n): `,

	"progress": `
{{separator 60}}
{{center "Installing" 60}}
{{separator 60}}

{{progressBar .Progress}} {{.Progress}}%
{{.ProgressText}}

Please wait while {{.AppName}} is being installed...`,

	"complete": `
{{separator 60}}
{{center "Installation Complete" 60}}
{{separator 60}}

{{.AppName}} has been successfully installed!

Installation location: {{.InstallPath}}
Total size: {{.TotalSize}}

Thank you for installing {{.AppName}}!

Press Enter to exit...`,
}

// getEmbeddedHTMLTemplate returns embedded HTML template
func getEmbeddedHTMLTemplate(name string) *template.Template {
	if tmplStr, exists := htmlTemplates[name]; exists {
		return template.Must(template.New(name).Funcs(htmlFuncMap()).Parse(tmplStr))
	}
	return template.Must(template.New(name).Parse("<html><body><h1>Template not found: {{.}}</h1></body></html>"))
}

// getEmbeddedCLITemplate returns embedded CLI template
func getEmbeddedCLITemplate(name string) *texttemplate.Template {
	if tmplStr, exists := cliTemplates[name]; exists {
		return texttemplate.Must(texttemplate.New(name).Funcs(cliFuncMap()).Parse(tmplStr))
	}
	return texttemplate.Must(texttemplate.New(name).Parse("Template not found: {{.}}"))
}