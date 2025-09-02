// pkg/ui/controls/controls.go
package controls

import (
    "fmt"
    "strings"
)

// ControlType definiert die verschiedenen Control-Typen
type ControlType int

const (
    TextInput ControlType = iota
    PasswordInput
    NumberInput
    Select
    Checkbox
    Button
    ProgressBar
    TextArea
)

// Control Interface für alle UI-Controls
type Control interface {
    GetType() ControlType
    GetName() string
    GetValue() interface{}
    SetValue(interface{}) error
    Validate() error
    Render(renderer Renderer) (string, error)
}

// Renderer Interface für verschiedene Output-Formate
type Renderer interface {
    RenderTextInput(ctrl *TextInputControl) (string, error)
    RenderPasswordInput(ctrl *PasswordInputControl) (string, error)
    RenderNumberInput(ctrl *NumberInputControl) (string, error)
    RenderSelect(ctrl *SelectControl) (string, error)
    RenderCheckbox(ctrl *CheckboxControl) (string, error)
    RenderButton(ctrl *ButtonControl) (string, error)
    RenderProgress(ctrl *ProgressControl) (string, error)
    RenderTextArea(ctrl *TextAreaControl) (string, error)
}

// TextInputControl für Texteingabe
type TextInputControl struct {
    name        string
    label       string
    value       string
    placeholder string
    required    bool
    validator   func(string) error
    cssClass    string
}

func NewTextInput(name, label string) *TextInputControl {
    return &TextInputControl{
        name:  name,
        label: label,
    }
}

func (t *TextInputControl) SetPlaceholder(placeholder string) *TextInputControl {
    t.placeholder = placeholder
    return t
}

func (t *TextInputControl) SetRequired(required bool) *TextInputControl {
    t.required = required
    return t
}

func (t *TextInputControl) SetValidator(validator func(string) error) *TextInputControl {
    t.validator = validator
    return t
}

func (t *TextInputControl) SetCSSClass(class string) *TextInputControl {
    t.cssClass = class
    return t
}

func (t *TextInputControl) GetType() ControlType { return TextInput }
func (t *TextInputControl) GetName() string     { return t.name }
func (t *TextInputControl) GetValue() interface{} { return t.value }

func (t *TextInputControl) SetValue(value interface{}) error {
    if str, ok := value.(string); ok {
        t.value = str
        return nil
    }
    return fmt.Errorf("expected string value for text input")
}

func (t *TextInputControl) Validate() error {
    if t.required && t.value == "" {
        return fmt.Errorf("%s is required", t.label)
    }
    if t.validator != nil {
        return t.validator(t.value)
    }
    return nil
}

func (t *TextInputControl) Render(renderer Renderer) (string, error) {
    return renderer.RenderTextInput(t)
}

// NumberInputControl für Zahleneingabe
type NumberInputControl struct {
    name      string
    label     string
    value     int
    min       *int
    max       *int
    required  bool
    validator func(int) error
}

func NewNumberInput(name, label string) *NumberInputControl {
    return &NumberInputControl{
        name:  name,
        label: label,
    }
}

func (n *NumberInputControl) SetMin(min int) *NumberInputControl {
    n.min = &min
    return n
}

func (n *NumberInputControl) SetMax(max int) *NumberInputControl {
    n.max = &max
    return n
}

func (n *NumberInputControl) SetRequired(required bool) *NumberInputControl {
    n.required = required
    return n
}

func (n *NumberInputControl) GetType() ControlType { return NumberInput }
func (n *NumberInputControl) GetName() string     { return n.name }
func (n *NumberInputControl) GetValue() interface{} { return n.value }

func (n *NumberInputControl) SetValue(value interface{}) error {
    if num, ok := value.(int); ok {
        n.value = num
        return nil
    }
    return fmt.Errorf("expected int value for number input")
}

func (n *NumberInputControl) Validate() error {
    if n.required && n.value == 0 {
        return fmt.Errorf("%s is required", n.label)
    }
    if n.min != nil && n.value < *n.min {
        return fmt.Errorf("%s must be at least %d", n.label, *n.min)
    }
    if n.max != nil && n.value > *n.max {
        return fmt.Errorf("%s must be at most %d", n.label, *n.max)
    }
    if n.validator != nil {
        return n.validator(n.value)
    }
    return nil
}

func (n *NumberInputControl) Render(renderer Renderer) (string, error) {
    return renderer.RenderNumberInput(n)
}

// SelectControl für Dropdown-Auswahl
type SelectControl struct {
    name      string
    label     string
    value     string
    options   []Option
    required  bool
    multiple  bool
    validator func(string) error
}

type Option struct {
    Value string
    Label string
}

func NewSelect(name, label string, options []Option) *SelectControl {
    return &SelectControl{
        name:    name,
        label:   label,
        options: options,
    }
}

func (s *SelectControl) SetRequired(required bool) *SelectControl {
    s.required = required
    return s
}

func (s *SelectControl) SetMultiple(multiple bool) *SelectControl {
    s.multiple = multiple
    return s
}

func (s *SelectControl) GetType() ControlType { return Select }
func (s *SelectControl) GetName() string     { return s.name }
func (s *SelectControl) GetValue() interface{} { return s.value }

func (s *SelectControl) SetValue(value interface{}) error {
    if str, ok := value.(string); ok {
        s.value = str
        return nil
    }
    return fmt.Errorf("expected string value for select")
}

func (s *SelectControl) Validate() error {
    if s.required && s.value == "" {
        return fmt.Errorf("%s is required", s.label)
    }
    if s.validator != nil {
        return s.validator(s.value)
    }
    return nil
}

func (s *SelectControl) Render(renderer Renderer) (string, error) {
    return renderer.RenderSelect(s)
}

// CheckboxControl für Boolean-Werte
type CheckboxControl struct {
    name      string
    label     string
    value     bool
    required  bool
    validator func(bool) error
}

func NewCheckbox(name, label string) *CheckboxControl {
    return &CheckboxControl{
        name:  name,
        label: label,
    }
}

func (c *CheckboxControl) SetRequired(required bool) *CheckboxControl {
    c.required = required
    return c
}

func (c *CheckboxControl) GetType() ControlType { return Checkbox }
func (c *CheckboxControl) GetName() string     { return c.name }
func (c *CheckboxControl) GetValue() interface{} { return c.value }

func (c *CheckboxControl) SetValue(value interface{}) error {
    if b, ok := value.(bool); ok {
        c.value = b
        return nil
    }
    return fmt.Errorf("expected bool value for checkbox")
}

func (c *CheckboxControl) Validate() error {
    if c.required && !c.value {
        return fmt.Errorf("%s must be checked", c.label)
    }
    if c.validator != nil {
        return c.validator(c.value)
    }
    return nil
}

func (c *CheckboxControl) Render(renderer Renderer) (string, error) {
    return renderer.RenderCheckbox(c)
}

// PasswordInputControl für Passwort-Eingabe
type PasswordInputControl struct {
    name      string
    label     string
    value     string
    required  bool
    minLength int
    validator func(string) error
}

func NewPasswordInput(name, label string) *PasswordInputControl {
    return &PasswordInputControl{
        name:  name,
        label: label,
    }
}

func (p *PasswordInputControl) SetRequired(required bool) *PasswordInputControl {
    p.required = required
    return p
}

func (p *PasswordInputControl) SetMinLength(length int) *PasswordInputControl {
    p.minLength = length
    return p
}

func (p *PasswordInputControl) GetType() ControlType { return PasswordInput }
func (p *PasswordInputControl) GetName() string     { return p.name }
func (p *PasswordInputControl) GetValue() interface{} { return p.value }

func (p *PasswordInputControl) SetValue(value interface{}) error {
    if str, ok := value.(string); ok {
        p.value = str
        return nil
    }
    return fmt.Errorf("expected string value for password input")
}

func (p *PasswordInputControl) Validate() error {
    if p.required && p.value == "" {
        return fmt.Errorf("%s is required", p.label)
    }
    if len(p.value) < p.minLength {
        return fmt.Errorf("%s must be at least %d characters", p.label, p.minLength)
    }
    if p.validator != nil {
        return p.validator(p.value)
    }
    return nil
}

func (p *PasswordInputControl) Render(renderer Renderer) (string, error) {
    return renderer.RenderPasswordInput(p)
}

// ButtonControl für Aktionen
type ButtonControl struct {
    name    string
    label   string
    action  string
    variant string // primary, secondary, danger, etc.
}

func NewButton(name, label, action string) *ButtonControl {
    return &ButtonControl{
        name:    name,
        label:   label,
        action:  action,
        variant: "primary",
    }
}

func (b *ButtonControl) SetVariant(variant string) *ButtonControl {
    b.variant = variant
    return b
}

func (b *ButtonControl) GetType() ControlType { return Button }
func (b *ButtonControl) GetName() string     { return b.name }
func (b *ButtonControl) GetValue() interface{} { return b.action }
func (b *ButtonControl) SetValue(value interface{}) error { return nil }
func (b *ButtonControl) Validate() error { return nil }

func (b *ButtonControl) Render(renderer Renderer) (string, error) {
    return renderer.RenderButton(b)
}

// ProgressControl für Fortschrittsanzeige
type ProgressControl struct {
    name    string
    current int
    total   int
    message string
}

func NewProgress(name string, current, total int, message string) *ProgressControl {
    return &ProgressControl{
        name:    name,
        current: current,
        total:   total,
        message: message,
    }
}

func (p *ProgressControl) GetType() ControlType { return ProgressBar }
func (p *ProgressControl) GetName() string     { return p.name }
func (p *ProgressControl) GetValue() interface{} { return p.current }
func (p *ProgressControl) SetValue(value interface{}) error { return nil }
func (p *ProgressControl) Validate() error { return nil }

func (p *ProgressControl) Render(renderer Renderer) (string, error) {
    return renderer.RenderProgress(p)
}

// TextAreaControl für mehrzeilige Texteingabe
type TextAreaControl struct {
    name        string
    label       string
    value       string
    placeholder string
    rows        int
    required    bool
    validator   func(string) error
}

func NewTextArea(name, label string) *TextAreaControl {
    return &TextAreaControl{
        name:  name,
        label: label,
        rows:  3,
    }
}

func (t *TextAreaControl) SetRows(rows int) *TextAreaControl {
    t.rows = rows
    return t
}

func (t *TextAreaControl) SetPlaceholder(placeholder string) *TextAreaControl {
    t.placeholder = placeholder
    return t
}

func (t *TextAreaControl) SetRequired(required bool) *TextAreaControl {
    t.required = required
    return t
}

func (t *TextAreaControl) GetType() ControlType { return TextArea }
func (t *TextAreaControl) GetName() string     { return t.name }
func (t *TextAreaControl) GetValue() interface{} { return t.value }

func (t *TextAreaControl) SetValue(value interface{}) error {
    if str, ok := value.(string); ok {
        t.value = str
        return nil
    }
    return fmt.Errorf("expected string value for textarea")
}

func (t *TextAreaControl) Validate() error {
    if t.required && t.value == "" {
        return fmt.Errorf("%s is required", t.label)
    }
    if t.validator != nil {
        return t.validator(t.value)
    }
    return nil
}

func (t *TextAreaControl) Render(renderer Renderer) (string, error) {
    return renderer.RenderTextArea(t)
}

// HTMLRenderer für Web-Interface
type HTMLRenderer struct {
    theme string
}

func NewHTMLRenderer(theme string) *HTMLRenderer {
    return &HTMLRenderer{theme: theme}
}

func (h *HTMLRenderer) RenderTextInput(ctrl *TextInputControl) (string, error) {
    class := "form-control"
    if ctrl.cssClass != "" {
        class += " " + ctrl.cssClass
    }
    
    html := fmt.Sprintf(`
        <div class="form-group">
            <label for="%s">%s%s</label>
            <input 
                type="text" 
                id="%s" 
                name="%s" 
                value="%s" 
                placeholder="%s"
                class="%s"
                %s
            />
        </div>`,
        ctrl.name, 
        ctrl.label,
        ternary(ctrl.required, " <span class=\"required\">*</span>", ""),
        ctrl.name, 
        ctrl.name, 
        ctrl.value, 
        ctrl.placeholder,
        class,
        ternary(ctrl.required, "required", ""),
    )
    return html, nil
}

func (h *HTMLRenderer) RenderPasswordInput(ctrl *PasswordInputControl) (string, error) {
    html := fmt.Sprintf(`
        <div class="form-group">
            <label for="%s">%s%s</label>
            <input 
                type="password" 
                id="%s" 
                name="%s" 
                class="form-control"
                %s
                %s
            />
        </div>`,
        ctrl.name, 
        ctrl.label,
        ternary(ctrl.required, " <span class=\"required\">*</span>", ""),
        ctrl.name, 
        ctrl.name,
        ternary(ctrl.required, "required", ""),
        ternary(ctrl.minLength > 0, fmt.Sprintf("minlength=\"%d\"", ctrl.minLength), ""),
    )
    return html, nil
}

func (h *HTMLRenderer) RenderNumberInput(ctrl *NumberInputControl) (string, error) {
    minAttr := ""
    maxAttr := ""
    if ctrl.min != nil {
        minAttr = fmt.Sprintf("min=\"%d\"", *ctrl.min)
    }
    if ctrl.max != nil {
        maxAttr = fmt.Sprintf("max=\"%d\"", *ctrl.max)
    }
    
    html := fmt.Sprintf(`
        <div class="form-group">
            <label for="%s">%s%s</label>
            <input 
                type="number" 
                id="%s" 
                name="%s" 
                value="%d" 
                class="form-control"
                %s %s %s
            />
        </div>`,
        ctrl.name, 
        ctrl.label,
        ternary(ctrl.required, " <span class=\"required\">*</span>", ""),
        ctrl.name, 
        ctrl.name, 
        ctrl.value,
        ternary(ctrl.required, "required", ""),
        minAttr,
        maxAttr,
    )
    return html, nil
}

func (h *HTMLRenderer) RenderSelect(ctrl *SelectControl) (string, error) {
    options := ""
    for _, opt := range ctrl.options {
        selected := ""
        if opt.Value == ctrl.value {
            selected = "selected"
        }
        options += fmt.Sprintf(`<option value="%s" %s>%s</option>`, 
            opt.Value, selected, opt.Label)
    }
    
    multiple := ""
    if ctrl.multiple {
        multiple = "multiple"
    }
    
    html := fmt.Sprintf(`
        <div class="form-group">
            <label for="%s">%s%s</label>
            <select id="%s" name="%s" class="form-control" %s %s>
                %s
            </select>
        </div>`,
        ctrl.name, 
        ctrl.label,
        ternary(ctrl.required, " <span class=\"required\">*</span>", ""),
        ctrl.name, 
        ctrl.name,
        ternary(ctrl.required, "required", ""),
        multiple,
        options,
    )
    return html, nil
}

func (h *HTMLRenderer) RenderCheckbox(ctrl *CheckboxControl) (string, error) {
    checked := ""
    if ctrl.value {
        checked = "checked"
    }
    
    html := fmt.Sprintf(`
        <div class="form-group">
            <div class="form-check">
                <input 
                    type="checkbox" 
                    id="%s" 
                    name="%s" 
                    class="form-check-input"
                    %s %s
                />
                <label for="%s" class="form-check-label">
                    %s%s
                </label>
            </div>
        </div>`,
        ctrl.name, 
        ctrl.name,
        checked,
        ternary(ctrl.required, "required", ""),
        ctrl.name,
        ctrl.label,
        ternary(ctrl.required, " <span class=\"required\">*</span>", ""),
    )
    return html, nil
}

func (h *HTMLRenderer) RenderButton(ctrl *ButtonControl) (string, error) {
    class := fmt.Sprintf("btn btn-%s", ctrl.variant)
    
    html := fmt.Sprintf(`
        <button type="button" class="%s" onclick="%s">
            %s
        </button>`,
        class,
        ctrl.action,
        ctrl.label,
    )
    return html, nil
}

func (h *HTMLRenderer) RenderProgress(ctrl *ProgressControl) (string, error) {
    percentage := 0
    if ctrl.total > 0 {
        percentage = (ctrl.current * 100) / ctrl.total
    }
    
    html := fmt.Sprintf(`
        <div class="form-group">
            <div class="progress-wrapper">
                <div class="progress-label">%s</div>
                <div class="progress">
                    <div class="progress-bar" style="width: %d%%"></div>
                </div>
                <div class="progress-text">%d%%</div>
            </div>
        </div>`,
        ctrl.message,
        percentage,
        percentage,
    )
    return html, nil
}

func (h *HTMLRenderer) RenderTextArea(ctrl *TextAreaControl) (string, error) {
    html := fmt.Sprintf(`
        <div class="form-group">
            <label for="%s">%s%s</label>
            <textarea 
                id="%s" 
                name="%s" 
                rows="%d"
                placeholder="%s"
                class="form-control"
                %s
            >%s</textarea>
        </div>`,
        ctrl.name, 
        ctrl.label,
        ternary(ctrl.required, " <span class=\"required\">*</span>", ""),
        ctrl.name, 
        ctrl.name, 
        ctrl.rows,
        ctrl.placeholder,
        ternary(ctrl.required, "required", ""),
        ctrl.value,
    )
    return html, nil
}

// CLIRenderer für Terminal-Interface
type CLIRenderer struct {
    colors bool
}

func NewCLIRenderer(colors bool) *CLIRenderer {
    return &CLIRenderer{colors: colors}
}

func (c *CLIRenderer) RenderTextInput(ctrl *TextInputControl) (string, error) {
    prompt := ctrl.label
    if ctrl.required {
        prompt += " *"
    }
    if ctrl.placeholder != "" {
        prompt += fmt.Sprintf(" (%s)", ctrl.placeholder)
    }
    prompt += ": "
    
    if c.colors {
        if ctrl.required {
            prompt = colorize(prompt, "yellow")
        }
    }
    
    return prompt, nil
}

func (c *CLIRenderer) RenderPasswordInput(ctrl *PasswordInputControl) (string, error) {
    prompt := ctrl.label
    if ctrl.required {
        prompt += " *"
    }
    if ctrl.minLength > 0 {
        prompt += fmt.Sprintf(" (min %d chars)", ctrl.minLength)
    }
    prompt += ": "
    
    if c.colors {
        prompt = colorize(prompt, "cyan")
    }
    
    return prompt, nil
}

func (c *CLIRenderer) RenderNumberInput(ctrl *NumberInputControl) (string, error) {
    prompt := ctrl.label
    
    if ctrl.min != nil && ctrl.max != nil {
        prompt += fmt.Sprintf(" (%d-%d)", *ctrl.min, *ctrl.max)
    } else if ctrl.min != nil {
        prompt += fmt.Sprintf(" (min %d)", *ctrl.min)
    } else if ctrl.max != nil {
        prompt += fmt.Sprintf(" (max %d)", *ctrl.max)
    }
    
    if ctrl.required {
        prompt += " *"
    }
    prompt += ": "
    
    return prompt, nil
}

func (c *CLIRenderer) RenderSelect(ctrl *SelectControl) (string, error) {
    prompt := fmt.Sprintf("%s:\n", ctrl.label)
    
    for i, opt := range ctrl.options {
        marker := " "
        if opt.Value == ctrl.value {
            marker = ">"
        }
        prompt += fmt.Sprintf("  %s %d) %s\n", marker, i+1, opt.Label)
    }
    prompt += "Choice: "
    
    return prompt, nil
}

func (c *CLIRenderer) RenderCheckbox(ctrl *CheckboxControl) (string, error) {
    marker := "[ ]"
    if ctrl.value {
        marker = "[x]"
    }
    
    prompt := fmt.Sprintf("%s %s", marker, ctrl.label)
    if ctrl.required {
        prompt += " *"
    }
    prompt += " [y/N]: "
    
    if c.colors {
        prompt = colorize(prompt, "cyan")
    }
    
    return prompt, nil
}

func (c *CLIRenderer) RenderButton(ctrl *ButtonControl) (string, error) {
    button := fmt.Sprintf("[%s]", ctrl.label)
    
    if c.colors {
        switch ctrl.variant {
        case "primary":
            button = colorize(button, "blue")
        case "danger":
            button = colorize(button, "red")
        case "secondary":
            button = colorize(button, "white")
        }
    }
    
    return button, nil
}

func (c *CLIRenderer) RenderProgress(ctrl *ProgressControl) (string, error) {
    percentage := 0
    if ctrl.total > 0 {
        percentage = (ctrl.current * 100) / ctrl.total
    }
    
    progressBar := createProgressBar(ctrl.current, ctrl.total, 40)
    
    output := fmt.Sprintf("%s %d%% - %s", progressBar, percentage, ctrl.message)
    
    if c.colors {
        output = colorize(output, "green")
    }
    
    return output, nil
}

func (c *CLIRenderer) RenderTextArea(ctrl *TextAreaControl) (string, error) {
    prompt := ctrl.label
    if ctrl.required {
        prompt += " *"
    }
    if ctrl.placeholder != "" {
        prompt += fmt.Sprintf(" (%s)", ctrl.placeholder)
    }
    prompt += fmt.Sprintf(" (Enter %d lines):\n", ctrl.rows)
    
    return prompt, nil
}

// Utility-Funktionen
func ternary(condition bool, trueVal, falseVal string) string {
    if condition {
        return trueVal
    }
    return falseVal
}

func colorize(text, color string) string {
    colors := map[string]string{
        "red":    "\033[31m",
        "green":  "\033[32m",
        "yellow": "\033[33m",
        "blue":   "\033[34m",
        "purple": "\033[35m",
        "cyan":   "\033[36m",
        "white":  "\033[37m",
    }
    
    if code, exists := colors[color]; exists {
        return fmt.Sprintf("%s%s\033[0m", code, text)
    }
    
    return text
}

func createProgressBar(current, total, width int) string {
    if total == 0 {
        return ""
    }
    
    filled := int(float64(current) / float64(total) * float64(width))
    empty := width - filled
    
    return "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"
}
