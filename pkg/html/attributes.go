// Package html - Extended attribute system for HTML elements
package html

import (
	"fmt"
	"strings"
)

// Common HTML Attributes

// Title sets the title attribute (tooltip)
func (e *Element) Title(title string) *Element {
	return e.Attr("title", title)
}

// Lang sets the lang attribute (language)
func (e *Element) Lang(lang string) *Element {
	return e.Attr("lang", lang)
}

// Dir sets the dir attribute (text direction: ltr, rtl, auto)
func (e *Element) Dir(dir string) *Element {
	return e.Attr("dir", dir)
}

// TabIndex sets the tabindex attribute
func (e *Element) TabIndex(index int) *Element {
	return e.Attr("tabindex", fmt.Sprintf("%d", index))
}

// AccessKey sets the accesskey attribute (keyboard shortcut)
func (e *Element) AccessKey(key string) *Element {
	return e.Attr("accesskey", key)
}

// Hidden sets the hidden attribute
func (e *Element) Hidden() *Element {
	return e.Attr("hidden", "hidden")
}

// Draggable sets the draggable attribute
func (e *Element) Draggable(draggable bool) *Element {
	if draggable {
		return e.Attr("draggable", "true")
	}
	return e.Attr("draggable", "false")
}

// Spellcheck sets the spellcheck attribute
func (e *Element) Spellcheck(spellcheck bool) *Element {
	if spellcheck {
		return e.Attr("spellcheck", "true")
	}
	return e.Attr("spellcheck", "false")
}

// Class Management

// AddClass adds a class to the existing class list
func (e *Element) AddClass(class string) *Element {
	existing := e.attributes["class"]
	if existing == "" {
		return e.Class(class)
	}
	classes := strings.Fields(existing)
	// Check if class already exists
	for _, c := range classes {
		if c == class {
			return e // Already exists
		}
	}
	classes = append(classes, class)
	return e.Class(strings.Join(classes, " "))
}

// RemoveClass removes a class from the class list
func (e *Element) RemoveClass(class string) *Element {
	existing := e.attributes["class"]
	if existing == "" {
		return e
	}
	classes := strings.Fields(existing)
	filtered := make([]string, 0, len(classes))
	for _, c := range classes {
		if c != class {
			filtered = append(filtered, c)
		}
	}
	if len(filtered) == 0 {
		delete(e.attributes, "class")
	} else {
		e.attributes["class"] = strings.Join(filtered, " ")
	}
	return e
}

// ToggleClass toggles a class in the class list
func (e *Element) ToggleClass(class string) *Element {
	existing := e.attributes["class"]
	classes := strings.Fields(existing)
	found := false
	for i, c := range classes {
		if c == class {
			// Remove it
			classes = append(classes[:i], classes[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		classes = append(classes, class)
	}
	if len(classes) == 0 {
		delete(e.attributes, "class")
	} else {
		e.attributes["class"] = strings.Join(classes, " ")
	}
	return e
}

// HasClass checks if the element has a specific class
func (e *Element) HasClass(class string) bool {
	existing := e.attributes["class"]
	if existing == "" {
		return false
	}
	classes := strings.Fields(existing)
	for _, c := range classes {
		if c == class {
			return true
		}
	}
	return false
}

// Event Handlers

// OnClick sets the onclick event handler
func (e *Element) OnClick(handler string) *Element {
	return e.Attr("onclick", handler)
}

// OnChange sets the onchange event handler
func (e *Element) OnChange(handler string) *Element {
	return e.Attr("onchange", handler)
}

// OnSubmit sets the onsubmit event handler
func (e *Element) OnSubmit(handler string) *Element {
	return e.Attr("onsubmit", handler)
}

// OnLoad sets the onload event handler
func (e *Element) OnLoad(handler string) *Element {
	return e.Attr("onload", handler)
}

// OnMouseOver sets the onmouseover event handler
func (e *Element) OnMouseOver(handler string) *Element {
	return e.Attr("onmouseover", handler)
}

// OnMouseOut sets the onmouseout event handler
func (e *Element) OnMouseOut(handler string) *Element {
	return e.Attr("onmouseout", handler)
}

// OnMouseEnter sets the onmouseenter event handler
func (e *Element) OnMouseEnter(handler string) *Element {
	return e.Attr("onmouseenter", handler)
}

// OnMouseLeave sets the onmouseleave event handler
func (e *Element) OnMouseLeave(handler string) *Element {
	return e.Attr("onmouseleave", handler)
}

// OnFocus sets the onfocus event handler
func (e *Element) OnFocus(handler string) *Element {
	return e.Attr("onfocus", handler)
}

// OnBlur sets the onblur event handler
func (e *Element) OnBlur(handler string) *Element {
	return e.Attr("onblur", handler)
}

// OnKeyDown sets the onkeydown event handler
func (e *Element) OnKeyDown(handler string) *Element {
	return e.Attr("onkeydown", handler)
}

// OnKeyUp sets the onkeyup event handler
func (e *Element) OnKeyUp(handler string) *Element {
	return e.Attr("onkeyup", handler)
}

// OnKeyPress sets the onkeypress event handler
func (e *Element) OnKeyPress(handler string) *Element {
	return e.Attr("onkeypress", handler)
}

// Data Attributes

// Data sets a data-* attribute
func (e *Element) Data(name, value string) *Element {
	return e.Attr("data-"+name, value)
}

// DataToggle sets data-toggle attribute (commonly used with Bootstrap)
func (e *Element) DataToggle(value string) *Element {
	return e.Data("toggle", value)
}

// DataTarget sets data-target attribute (commonly used with Bootstrap)
func (e *Element) DataTarget(value string) *Element {
	return e.Data("target", value)
}

// DataDismiss sets data-dismiss attribute (commonly used with Bootstrap)
func (e *Element) DataDismiss(value string) *Element {
	return e.Data("dismiss", value)
}

// DataPlacement sets data-placement attribute (commonly used with Bootstrap tooltips)
func (e *Element) DataPlacement(value string) *Element {
	return e.Data("placement", value)
}

// ARIA Attributes for Accessibility

// Role sets the role attribute for accessibility
func (e *Element) Role(role string) *Element {
	return e.Attr("role", role)
}

// AriaLabel sets the aria-label attribute
func (e *Element) AriaLabel(label string) *Element {
	return e.Attr("aria-label", label)
}

// AriaLabelledBy sets the aria-labelledby attribute
func (e *Element) AriaLabelledBy(id string) *Element {
	return e.Attr("aria-labelledby", id)
}

// AriaDescribedBy sets the aria-describedby attribute
func (e *Element) AriaDescribedBy(id string) *Element {
	return e.Attr("aria-describedby", id)
}

// AriaHidden sets the aria-hidden attribute
func (e *Element) AriaHidden(hidden bool) *Element {
	if hidden {
		return e.Attr("aria-hidden", "true")
	}
	return e.Attr("aria-hidden", "false")
}

// AriaExpanded sets the aria-expanded attribute
func (e *Element) AriaExpanded(expanded bool) *Element {
	if expanded {
		return e.Attr("aria-expanded", "true")
	}
	return e.Attr("aria-expanded", "false")
}

// AriaPressed sets the aria-pressed attribute
func (e *Element) AriaPressed(pressed bool) *Element {
	if pressed {
		return e.Attr("aria-pressed", "true")
	}
	return e.Attr("aria-pressed", "false")
}

// AriaChecked sets the aria-checked attribute
func (e *Element) AriaChecked(checked bool) *Element {
	if checked {
		return e.Attr("aria-checked", "true")
	}
	return e.Attr("aria-checked", "false")
}

// AriaSelected sets the aria-selected attribute
func (e *Element) AriaSelected(selected bool) *Element {
	if selected {
		return e.Attr("aria-selected", "true")
	}
	return e.Attr("aria-selected", "false")
}

// AriaDisabled sets the aria-disabled attribute
func (e *Element) AriaDisabled(disabled bool) *Element {
	if disabled {
		return e.Attr("aria-disabled", "true")
	}
	return e.Attr("aria-disabled", "false")
}

// AriaCurrent sets the aria-current attribute
func (e *Element) AriaCurrent(current string) *Element {
	return e.Attr("aria-current", current) // "page", "step", "location", "date", "time", "true", "false"
}

// AriaLive sets the aria-live attribute
func (e *Element) AriaLive(live string) *Element {
	return e.Attr("aria-live", live) // "off", "polite", "assertive"
}

// AriaControls sets the aria-controls attribute
func (e *Element) AriaControls(controls string) *Element {
	return e.Attr("aria-controls", controls)
}

// Form-specific Attributes

// Name sets the name attribute (for form elements)
func (e *Element) Name(name string) *Element {
	return e.Attr("name", name)
}

// Value sets the value attribute (for form elements)
func (e *Element) Value(value string) *Element {
	return e.Attr("value", value)
}

// Placeholder sets the placeholder attribute
func (e *Element) Placeholder(placeholder string) *Element {
	return e.Attr("placeholder", placeholder)
}

// Required sets the required attribute
func (e *Element) Required() *Element {
	return e.Attr("required", "required")
}

// Disabled sets the disabled attribute
func (e *Element) Disabled() *Element {
	return e.Attr("disabled", "disabled")
}

// ReadOnly sets the readonly attribute
func (e *Element) ReadOnly() *Element {
	return e.Attr("readonly", "readonly")
}

// Checked sets the checked attribute (for checkboxes and radio buttons)
func (e *Element) Checked() *Element {
	return e.Attr("checked", "checked")
}

// Selected sets the selected attribute (for option elements)
func (e *Element) Selected() *Element {
	return e.Attr("selected", "selected")
}

// Multiple sets the multiple attribute (for select elements)
func (e *Element) Multiple() *Element {
	return e.Attr("multiple", "multiple")
}

// AutoComplete sets the autocomplete attribute
func (e *Element) AutoComplete(value string) *Element {
	return e.Attr("autocomplete", value)
}

// AutoFocus sets the autofocus attribute
func (e *Element) AutoFocus() *Element {
	return e.Attr("autofocus", "autofocus")
}

// MaxLength sets the maxlength attribute
func (e *Element) MaxLength(length int) *Element {
	return e.Attr("maxlength", fmt.Sprintf("%d", length))
}

// MinLength sets the minlength attribute
func (e *Element) MinLength(length int) *Element {
	return e.Attr("minlength", fmt.Sprintf("%d", length))
}

// Min sets the min attribute (for number/date inputs)
func (e *Element) Min(min string) *Element {
	return e.Attr("min", min)
}

// Max sets the max attribute (for number/date inputs)
func (e *Element) Max(max string) *Element {
	return e.Attr("max", max)
}

// Step sets the step attribute (for number inputs)
func (e *Element) Step(step string) *Element {
	return e.Attr("step", step)
}

// Pattern sets the pattern attribute (for input validation)
func (e *Element) Pattern(pattern string) *Element {
	return e.Attr("pattern", pattern)
}

// Link-specific Attributes

// Href sets the href attribute (for links)
func (e *Element) Href(href string) *Element {
	return e.Attr("href", href)
}

// Target sets the target attribute (for links)
func (e *Element) Target(target string) *Element {
	return e.Attr("target", target)
}

// Rel sets the rel attribute (for links)
func (e *Element) Rel(rel string) *Element {
	return e.Attr("rel", rel)
}

// Download sets the download attribute (for links)
func (e *Element) Download(filename ...string) *Element {
	if len(filename) > 0 {
		return e.Attr("download", filename[0])
	}
	return e.Attr("download", "")
}

// Image-specific Attributes

// Src sets the src attribute (for images, scripts, etc.)
func (e *Element) Src(src string) *Element {
	return e.Attr("src", src)
}

// Alt sets the alt attribute (for images)
func (e *Element) Alt(alt string) *Element {
	return e.Attr("alt", alt)
}

// Width sets the width attribute
func (e *Element) Width(width string) *Element {
	return e.Attr("width", width)
}

// Height sets the height attribute
func (e *Element) Height(height string) *Element {
	return e.Attr("height", height)
}

// Loading sets the loading attribute (for images)
func (e *Element) Loading(loading string) *Element {
	return e.Attr("loading", loading) // "lazy", "eager"
}

// Table-specific Attributes

// Colspan sets the colspan attribute (for table cells)
func (e *Element) Colspan(span int) *Element {
	return e.Attr("colspan", fmt.Sprintf("%d", span))
}

// Rowspan sets the rowspan attribute (for table cells)
func (e *Element) Rowspan(span int) *Element {
	return e.Attr("rowspan", fmt.Sprintf("%d", span))
}

// Scope sets the scope attribute (for table headers)
func (e *Element) Scope(scope string) *Element {
	return e.Attr("scope", scope) // "col", "row", "colgroup", "rowgroup"
}

// Media-specific Attributes

// Controls sets the controls attribute (for audio/video)
func (e *Element) Controls() *Element {
	return e.Attr("controls", "controls")
}

// AutoPlay sets the autoplay attribute (for audio/video)
func (e *Element) AutoPlay() *Element {
	return e.Attr("autoplay", "autoplay")
}

// Loop sets the loop attribute (for audio/video)
func (e *Element) Loop() *Element {
	return e.Attr("loop", "loop")
}

// Muted sets the muted attribute (for audio/video)
func (e *Element) Muted() *Element {
	return e.Attr("muted", "muted")
}

// Preload sets the preload attribute (for audio/video)
func (e *Element) Preload(preload string) *Element {
	return e.Attr("preload", preload) // "none", "metadata", "auto"
}

// Poster sets the poster attribute (for video)
func (e *Element) Poster(poster string) *Element {
	return e.Attr("poster", poster)
}

// Content Security and Meta Attributes

// ContentType sets the content-type for meta elements
func (e *Element) ContentType(contentType string) *Element {
	return e.Attr("content", contentType)
}

// Charset sets the charset attribute
func (e *Element) Charset(charset string) *Element {
	return e.Attr("charset", charset)
}

// HttpEquiv sets the http-equiv attribute (for meta elements)
func (e *Element) HttpEquiv(equiv string) *Element {
	return e.Attr("http-equiv", equiv)
}

// Content sets the content attribute (for meta elements)
func (e *Element) Content(content string) *Element {
	return e.Attr("content", content)
}

// Method sets the method attribute (for forms)
func (e *Element) Method(method string) *Element {
	return e.Attr("method", method) // "get", "post"
}

// Action sets the action attribute (for forms)
func (e *Element) Action(action string) *Element {
	return e.Attr("action", action)
}

// EncType sets the enctype attribute (for forms)
func (e *Element) EncType(enctype string) *Element {
	return e.Attr("enctype", enctype)
}

// NoValidate sets the novalidate attribute (for forms)
func (e *Element) NoValidate() *Element {
	return e.Attr("novalidate", "novalidate")
}

// Convenience Methods for Common Patterns

// OpenInNewTab sets target="_blank" and rel="noopener noreferrer" for security
func (e *Element) OpenInNewTab() *Element {
	return e.Target("_blank").Rel("noopener noreferrer")
}

// Tooltip sets title attribute for simple tooltip
func (e *Element) Tooltip(text string) *Element {
	return e.Title(text)
}

// BootstrapButton applies common Bootstrap button classes and attributes
func (e *Element) BootstrapButton(variant string) *Element {
	return e.AddClass("btn").AddClass("btn-" + variant)
}

// BootstrapModal applies common Bootstrap modal attributes
func (e *Element) BootstrapModal(modalId string) *Element {
	return e.DataToggle("modal").DataTarget("#" + modalId)
}

// BootstrapTooltip applies Bootstrap tooltip attributes
func (e *Element) BootstrapTooltip(text string) *Element {
	return e.DataToggle("tooltip").Title(text)
}

// SetIf conditionally sets an attribute based on a condition
func (e *Element) SetIf(condition bool, name, value string) *Element {
	if condition {
		return e.Attr(name, value)
	}
	return e
}

// AddClassIf conditionally adds a class based on a condition
func (e *Element) AddClassIf(condition bool, class string) *Element {
	if condition {
		return e.AddClass(class)
	}
	return e
}