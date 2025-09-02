// pkg/ui/interfaces.go
package ui

import (
    "bufio"
    "fmt"
    "io"
    "net/http"
    "strconv"
    "strings"
    "html/template"
)

// UserInterface definiert das Interface für alle UI-Implementierungen
type UserInterface interface {
    ShowStep(step Step) error
    GetUserInput() (Input, error) 
    ShowProgress(progress Progress) error
    ShowError(err error) error
    ShowForm(form *Form) error
}

// Step, Input, Progress von wizard package
type Step struct {
    Title       string
    Description string
    Fields      []Field
}

type Field struct {
    Name        string
    Type        string
    Label       string
    Required    bool
    Default     interface{}
    Options     []Option
    Placeholder string
}

type Option struct {
    Value string
    Label string
}

type Input map[string]interface{}

type Progress struct {
    Current int
    Total   int
    Message string
}

// CLI Interface Implementation
type CLIInterface struct {
    reader io.Reader
    writer io.Writer
    scanner *bufio.Scanner
    colors bool
}

func NewCLIInterface(reader io.Reader, writer io.Writer) *CLIInterface {
    return &CLIInterface{
        reader:  reader,
        writer:  writer,
        scanner: bufio.NewScanner(reader),
        colors:  true, // Terminal-Farben aktiviert
    }
}

func (cli *CLIInterface) ShowStep(step Step) error {
    cli.printHeader(step.Title)
    fmt.Fprintln(cli.writer, step.Description)
    fmt.Fprintln(cli.writer)
    
    return nil
}

func (cli *CLIInterface) GetUserInput() (Input, error) {
    input := make(Input)
    
    // Einfache Implementierung - liest eine Zeile
    if cli.scanner.Scan() {
        text := cli.scanner.Text()
        input["response"] = text
    }
    
    return input, cli.scanner.Err()
}

func (cli *CLIInterface) ShowProgress(progress Progress) error {
    percentage := float64(progress.Current) / float64(progress.Total) * 100
    
    progressBar := cli.createProgressBar(progress.Current, progress.Total, 50)
    
    if cli.colors {
        fmt.Fprintf(cli.writer, "\r\033[36m%s\033[0m %.1f%% - %s", 
            progressBar, percentage, progress.Message)
    } else {
        fmt.Fprintf(cli.writer, "\r%s %.1f%% - %s", 
            progressBar, percentage, progress.Message)
    }
    
    if progress.Current == progress.Total {
        fmt.Fprintln(cli.writer) // Neue Zeile bei Completion
    }
    
    return nil
}

func (cli *CLIInterface) ShowError(err error) error {
    if cli.colors {
        fmt.Fprintf(cli.writer, "\033[31mError: %s\033[0m\n", err.Error())
    } else {
        fmt.Fprintf(cli.writer, "Error: %s\n", err.Error())
    }
    return nil
}

func (cli *CLIInterface) ShowForm(form *Form) error {
    cli.printHeader(form.GetTitle())
    
    for _, field := range form.GetFields() {
        if err := cli.showField(field); err != nil {
            return err
        }
    }
    
    return nil
}

func (cli *CLIInterface) showField(field Field) error {
    switch field.Type {
    case "text", "password":
        return cli.showTextInput(field)
    case "select":
        return cli.showSelect(field)
    case "checkbox":
        return cli.showCheckbox(field)
    case "number":
        return cli.showNumberInput(field)
    }
    return nil
}

func (cli *CLIInterface) showTextInput(field Field) error {
    prompt := field.Label
    if field.Required {
        prompt += " *"
    }
    if field.Placeholder != "" {
        prompt += fmt.Sprintf(" (%s)", field.Placeholder)
    }
    prompt += ": "
    
    if cli.colors {
        if field.Required {
            prompt = cli.colorize(prompt, "yellow")
        }
    }
    
    fmt.Fprint(cli.writer, prompt)
    return nil
}

func (cli *CLIInterface) showSelect(field Field) error {
    fmt.Fprintf(cli.writer, "%s:\n", field.Label)
    
    for i, opt := range field.Options {
        fmt.Fprintf(cli.writer, "  %d) %s\n", i+1, opt.Label)
    }
    fmt.Fprint(cli.writer, "Choice: ")
    
    return nil
}

func (cli *CLIInterface) showCheckbox(field Field) error {
    prompt := fmt.Sprintf("%s [y/N]: ", field.Label)
    if cli.colors {
        prompt = cli.colorize(prompt, "cyan")
    }
    
    fmt.Fprint(cli.writer, prompt)
    return nil
}

func (cli *CLIInterface) showNumberInput(field Field) error {
    prompt := field.Label
    if field.Default != nil {
        prompt += fmt.Sprintf(" (default: %v)", field.Default)
    }
    prompt += ": "
    
    fmt.Fprint(cli.writer, prompt)
    return nil
}

func (cli *CLIInterface) printHeader(title string) {
    separator := strings.Repeat("=", len(title)+4)
    
    if cli.colors {
        fmt.Fprintf(cli.writer, "\n\033[1;34m%s\033[0m\n", separator)
        fmt.Fprintf(cli.writer, "\033[1;34m  %s\033[0m\n", title)
        fmt.Fprintf(cli.writer, "\033[1;34m%s\033[0m\n\n", separator)
    } else {
        fmt.Fprintf(cli.writer, "\n%s\n", separator)
        fmt.Fprintf(cli.writer, "  %s\n", title)
        fmt.Fprintf(cli.writer, "%s\n\n", separator)
    }
}

func (cli *CLIInterface) createProgressBar(current, total, width int) string {
    if total == 0 {
        return ""
    }
    
    filled := int(float64(current) / float64(total) * float64(width))
    empty := width - filled
    
    return "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"
}

func (cli *CLIInterface) colorize(text, color string) string {
    if !cli.colors {
        return text
    }
    
    colors := map[string]string{
        "red":    "31",
        "green":  "32", 
        "yellow": "33",
        "blue":   "34",
        "purple": "35",
        "cyan":   "36",
        "white":  "37",
    }
    
    if code, exists := colors[color]; exists {
        return fmt.Sprintf("\033[%sm%s\033[0m", code, text)
    }
    
    return text
}

// HTML/Web Interface Implementation
type HTMLInterface struct {
    server     *http.Server
    template   *template.Template
    currentForm *Form
    responseChannel chan Input
}

func NewHTMLInterface(addr string) *HTMLInterface {
    htmlUI := &HTMLInterface{
        responseChannel: make(chan Input),
    }
    
    mux := http.NewServeMux()
    mux.HandleFunc("/", htmlUI.handleIndex)
    mux.HandleFunc("/submit", htmlUI.handleSubmit)
    mux.HandleFunc("/assets/", htmlUI.handleAssets)
    
    htmlUI.server = &http.Server{
        Addr:    addr,
        Handler: mux,
    }
    
    // Template laden
    htmlUI.loadTemplates()
    
    return htmlUI
}

func (html *HTMLInterface) Start() error {
    return html.server.ListenAndServe()
}

func (html *HTMLInterface) ShowStep(step Step) error {
    // Step in Form konvertieren
    form := &Form{
        title: step.Title,
        description: step.Description,
        fields: step.Fields,
    }
    
    return html.ShowForm(form)
}

func (html *HTMLInterface) GetUserInput() (Input, error) {
    // Warten auf Form-Submit
    input := <-html.responseChannel
    return input, nil
}

func (html *HTMLInterface) ShowProgress(progress Progress) error {
    // Progress via WebSocket oder Server-Sent Events senden
    // Vereinfachte Implementierung hier
    return nil
}

func (html *HTMLInterface) ShowError(err error) error {
    // Error in Template einbetten
    return nil
}

func (html *HTMLInterface) ShowForm(form *Form) error {
    html.currentForm = form
    return nil
}

func (html *HTMLInterface) handleIndex(w http.ResponseWriter, r *http.Request) {
    if html.currentForm == nil {
        http.Error(w, "No form available", http.StatusNotFound)
        return
    }
    
    data := struct {
        Form *Form
        CSS  template.HTML
        JS   template.HTML
    }{
        Form: html.currentForm,
        CSS:  template.HTML(html.getCSS()),
        JS:   template.HTML(html.getJS()),
    }
    
    html.template.Execute(w, data)
}

func (html *HTMLInterface) handleSubmit(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    r.ParseForm()
    input := make(Input)
    
    for key, values := range r.Form {
        if len(values) > 0 {
            // Typ-Konvertierung basierend auf Feld-Typ
            input[key] = html.convertValue(key, values[0])
        }
    }
    
    // Input an GetUserInput() weiterleiten
    html.responseChannel <- input
    
    w.WriteHeader(http.StatusOK)
    fmt.Fprint(w, "OK")
}

func (html *HTMLInterface) convertValue(fieldName, value string) interface{} {
    // Finde Feld-Typ
    for _, field := range html.currentForm.fields {
        if field.Name == fieldName {
            switch field.Type {
            case "number":
                if num, err := strconv.Atoi(value); err == nil {
                    return num
                }
            case "checkbox":
                return value == "on"
            default:
                return value
            }
        }
    }
    return value
}

func (html *HTMLInterface) handleAssets(w http.ResponseWriter, r *http.Request) {
    // Statische Assets servieren (CSS, JS, Bilder)
    http.ServeFile(w, r, r.URL.Path[1:])
}

func (html *HTMLInterface) loadTemplates() {
    templateHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>{{.Form.Title}}</title>
    <style>{{.CSS}}</style>
</head>
<body>
    <div class="container">
        <h1>{{.Form.Title}}</h1>
        <p class="description">{{.Form.Description}}</p>
        
        <form id="wizardForm" action="/submit" method="POST">
            {{range .Form.Fields}}
            <div class="form-group">
                <label for="{{.Name}}">{{.Label}}{{if .Required}} *{{end}}</label>
                
                {{if eq .Type "text"}}
                <input type="text" id="{{.Name}}" name="{{.Name}}" 
                       placeholder="{{.Placeholder}}" {{if .Required}}required{{end}} 
                       class="form-control">
                       
                {{else if eq .Type "password"}}
                <input type="password" id="{{.Name}}" name="{{.Name}}" 
                       {{if .Required}}required{{end}} class="form-control">
                       
                {{else if eq .Type "number"}}
                <input type="number" id="{{.Name}}" name="{{.Name}}" 
                       {{if .Default}}value="{{.Default}}"{{end}}
                       {{if .Required}}required{{end}} class="form-control">
                       
                {{else if eq .Type "select"}}
                <select id="{{.Name}}" name="{{.Name}}" {{if .Required}}required{{end}} 
                        class="form-control">
                    {{range .Options}}
                    <option value="{{.Value}}">{{.Label}}</option>
                    {{end}}
                </select>
                
                {{else if eq .Type "checkbox"}}
                <input type="checkbox" id="{{.Name}}" name="{{.Name}}" 
                       {{if .Required}}required{{end}} class="form-check">
                {{end}}
            </div>
            {{end}}
            
            <div class="form-actions">
                <button type="button" onclick="goBack()" class="btn btn-secondary">Back</button>
                <button type="submit" class="btn btn-primary">Next</button>
            </div>
        </form>
    </div>
    
    <script>{{.JS}}</script>
</body>
</html>
`
    
    html.template = template.Must(template.New("wizard").Parse(templateHTML))
}

func (html *HTMLInterface) getCSS() string {
    return `
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; text-align: center; margin-bottom: 10px; }
        .description { color: #666; text-align: center; margin-bottom: 30px; }
        .form-group { margin-bottom: 20px; }
        label { display: block; font-weight: bold; margin-bottom: 5px; color: #333; }
        .form-control { width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 4px; font-size: 14px; }
        .form-control:focus { border-color: #007bff; box-shadow: 0 0 0 2px rgba(0,123,255,0.25); outline: none; }
        .form-check { margin-right: 10px; }
        .form-actions { text-align: center; margin-top: 30px; }
        .btn { padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; font-size: 14px; margin: 0 10px; }
        .btn-primary { background-color: #007bff; color: white; }
        .btn-secondary { background-color: #6c757d; color: white; }
        .btn:hover { opacity: 0.9; }
    `
}

func (html *HTMLInterface) getJS() string {
    return `
        function goBack() {
            // Implementierung für Zurück-Button
            alert('Going back...');
        }
        
        // Form-Validierung
        document.getElementById('wizardForm').addEventListener('submit', function(e) {
            const required = this.querySelectorAll('[required]');
            for (let field of required) {
                if (!field.value.trim()) {
                    alert('Please fill in all required fields');
                    e.preventDefault();
                    return false;
                }
            }
        });
        
        // Auto-Focus auf erstes Feld
        document.addEventListener('DOMContentLoaded', function() {
            const firstInput = document.querySelector('.form-control');
            if (firstInput) {
                firstInput.focus();
            }
        });
    `
}

// Form Wrapper für UI-Abstraktion
type Form struct {
    title       string
    description string
    fields      []Field
}

func (f *Form) GetTitle() string {
    return f.title
}

func (f *Form) GetDescription() string {
    return f.description
}

func (f *Form) GetFields() []Field {
    return f.fields
}

// Factory für UI-Erstellung
func CreateUI(mode string) (UserInterface, error) {
    switch mode {
    case "cli":
        return NewCLIInterface(nil, nil), nil // Reader/Writer werden später gesetzt
    case "html", "web":
        return NewHTMLInterface(":8080"), nil
    default:
        return nil, fmt.Errorf("unsupported UI mode: %s", mode)
    }
}
