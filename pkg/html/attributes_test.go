package html

import (
	"strings"
	"testing"
)

func TestBasicAttributes(t *testing.T) {
	tests := []struct {
		name     string
		element  *Element
		expected map[string]string // Expected attributes
	}{
		{
			name:     "ID and Title",
			element:  DIV().ID("test").Title("Test tooltip"),
			expected: map[string]string{"id": "test", "title": "Test tooltip"},
		},
		{
			name:     "TabIndex",
			element:  INPUT("text").TabIndex(1),
			expected: map[string]string{"type": "text", "tabindex": "1"},
		},
		{
			name:     "Hidden",
			element:  DIV().Hidden(),
			expected: map[string]string{"hidden": "hidden"},
		},
		{
			name:     "Draggable True",
			element:  DIV().Draggable(true),
			expected: map[string]string{"draggable": "true"},
		},
		{
			name:     "Draggable False",
			element:  DIV().Draggable(false),
			expected: map[string]string{"draggable": "false"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for expectedAttr, expectedValue := range tt.expected {
				actualValue, exists := tt.element.attributes[expectedAttr]
				if !exists {
					t.Errorf("Expected attribute %q not found", expectedAttr)
					continue
				}
				if actualValue != expectedValue {
					t.Errorf("For attribute %q: expected %q, got %q", expectedAttr, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestClassManagement(t *testing.T) {
	t.Run("AddClass to empty", func(t *testing.T) {
		element := DIV().AddClass("container")
		if element.attributes["class"] != "container" {
			t.Errorf("Expected class 'container', got %q", element.attributes["class"])
		}
	})

	t.Run("AddClass to existing", func(t *testing.T) {
		element := DIV().Class("existing").AddClass("new")
		expected := "existing new"
		if element.attributes["class"] != expected {
			t.Errorf("Expected class %q, got %q", expected, element.attributes["class"])
		}
	})

	t.Run("AddClass duplicate", func(t *testing.T) {
		element := DIV().Class("container").AddClass("container")
		if element.attributes["class"] != "container" {
			t.Errorf("Expected class 'container' (no duplicate), got %q", element.attributes["class"])
		}
	})

	t.Run("RemoveClass", func(t *testing.T) {
		element := DIV().Class("one two three").RemoveClass("two")
		expected := "one three"
		if element.attributes["class"] != expected {
			t.Errorf("Expected class %q, got %q", expected, element.attributes["class"])
		}
	})

	t.Run("RemoveClass all", func(t *testing.T) {
		element := DIV().Class("only").RemoveClass("only")
		_, exists := element.attributes["class"]
		if exists {
			t.Error("Expected class attribute to be removed")
		}
	})

	t.Run("ToggleClass add", func(t *testing.T) {
		element := DIV().Class("existing").ToggleClass("new")
		expected := "existing new"
		if element.attributes["class"] != expected {
			t.Errorf("Expected class %q, got %q", expected, element.attributes["class"])
		}
	})

	t.Run("ToggleClass remove", func(t *testing.T) {
		element := DIV().Class("existing remove").ToggleClass("remove")
		expected := "existing"
		if element.attributes["class"] != expected {
			t.Errorf("Expected class %q, got %q", expected, element.attributes["class"])
		}
	})

	t.Run("HasClass true", func(t *testing.T) {
		element := DIV().Class("one two three")
		if !element.HasClass("two") {
			t.Error("Expected HasClass('two') to be true")
		}
	})

	t.Run("HasClass false", func(t *testing.T) {
		element := DIV().Class("one three")
		if element.HasClass("two") {
			t.Error("Expected HasClass('two') to be false")
		}
	})
}

func TestEventHandlers(t *testing.T) {
	tests := []struct {
		name        string
		element     *Element
		expectedAttr string
		expectedVal  string
	}{
		{
			name:        "OnClick",
			element:     BUTTON("Click me").OnClick("handleClick()"),
			expectedAttr: "onclick",
			expectedVal:  "handleClick()",
		},
		{
			name:        "OnChange",
			element:     INPUT("text").OnChange("handleChange(this.value)"),
			expectedAttr: "onchange",
			expectedVal:  "handleChange(this.value)",
		},
		{
			name:        "OnSubmit",
			element:     FORM().OnSubmit("return validateForm()"),
			expectedAttr: "onsubmit",
			expectedVal:  "return validateForm()",
		},
		{
			name:        "OnMouseOver",
			element:     DIV().OnMouseOver("showTooltip()"),
			expectedAttr: "onmouseover",
			expectedVal:  "showTooltip()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualVal, exists := tt.element.attributes[tt.expectedAttr]
			if !exists {
				t.Errorf("Expected attribute %q not found", tt.expectedAttr)
				return
			}
			if actualVal != tt.expectedVal {
				t.Errorf("Expected %q, got %q", tt.expectedVal, actualVal)
			}
		})
	}
}

func TestDataAttributes(t *testing.T) {
	t.Run("Data attribute", func(t *testing.T) {
		element := DIV().Data("custom", "value")
		expected := "value"
		actual := element.attributes["data-custom"]
		if actual != expected {
			t.Errorf("Expected data-custom=%q, got %q", expected, actual)
		}
	})

	t.Run("Bootstrap data attributes", func(t *testing.T) {
		element := BUTTON("Toggle").
			DataToggle("modal").
			DataTarget("#myModal").
			DataDismiss("modal")

		tests := []struct {
			attr     string
			expected string
		}{
			{"data-toggle", "modal"},
			{"data-target", "#myModal"},
			{"data-dismiss", "modal"},
		}

		for _, test := range tests {
			actual := element.attributes[test.attr]
			if actual != test.expected {
				t.Errorf("Expected %s=%q, got %q", test.attr, test.expected, actual)
			}
		}
	})
}

func TestAriaAttributes(t *testing.T) {
	tests := []struct {
		name        string
		element     *Element
		expectedAttr string
		expectedVal  string
	}{
		{
			name:        "Role",
			element:     DIV().Role("button"),
			expectedAttr: "role",
			expectedVal:  "button",
		},
		{
			name:        "AriaLabel",
			element:     BUTTON("").AriaLabel("Close dialog"),
			expectedAttr: "aria-label",
			expectedVal:  "Close dialog",
		},
		{
			name:        "AriaHidden true",
			element:     SPAN("").AriaHidden(true),
			expectedAttr: "aria-hidden",
			expectedVal:  "true",
		},
		{
			name:        "AriaExpanded false",
			element:     BUTTON("").AriaExpanded(false),
			expectedAttr: "aria-expanded",
			expectedVal:  "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualVal := tt.element.attributes[tt.expectedAttr]
			if actualVal != tt.expectedVal {
				t.Errorf("Expected %q, got %q", tt.expectedVal, actualVal)
			}
		})
	}
}

func TestFormAttributes(t *testing.T) {
	input := INPUT("email").
		Name("user_email").
		Value("test@example.com").
		Placeholder("Enter your email").
		Required().
		MaxLength(100).
		Pattern("[a-z0-9._%+-]+@[a-z0-9.-]+\\.[a-z]{2,}$")

	tests := []struct {
		attr     string
		expected string
	}{
		{"type", "email"},
		{"name", "user_email"},
		{"value", "test@example.com"},
		{"placeholder", "Enter your email"},
		{"required", "required"},
		{"maxlength", "100"},
		{"pattern", "[a-z0-9._%+-]+@[a-z0-9.-]+\\.[a-z]{2,}$"},
	}

	for _, test := range tests {
		t.Run("Form "+test.attr, func(t *testing.T) {
			actual := input.attributes[test.attr]
			if actual != test.expected {
				t.Errorf("Expected %s=%q, got %q", test.attr, test.expected, actual)
			}
		})
	}
}

func TestLinkAttributes(t *testing.T) {
	link := A("https://example.com", "Visit").
		Target("_blank").
		Rel("noopener noreferrer").
		Download("file.pdf")

	tests := []struct {
		attr     string
		expected string
	}{
		{"href", "https://example.com"},
		{"target", "_blank"},
		{"rel", "noopener noreferrer"},
		{"download", "file.pdf"},
	}

	for _, test := range tests {
		t.Run("Link "+test.attr, func(t *testing.T) {
			actual := link.attributes[test.attr]
			if actual != test.expected {
				t.Errorf("Expected %s=%q, got %q", test.attr, test.expected, actual)
			}
		})
	}
}

func TestConvenienceMethods(t *testing.T) {
	t.Run("OpenInNewTab", func(t *testing.T) {
		link := A("#").OpenInNewTab()
		if link.attributes["target"] != "_blank" {
			t.Error("Expected target='_blank'")
		}
		if link.attributes["rel"] != "noopener noreferrer" {
			t.Error("Expected rel='noopener noreferrer'")
		}
	})

	t.Run("BootstrapButton", func(t *testing.T) {
		btn := BUTTON("Click").BootstrapButton("primary")
		class := btn.attributes["class"]
		if !strings.Contains(class, "btn") || !strings.Contains(class, "btn-primary") {
			t.Errorf("Expected Bootstrap button classes, got %q", class)
		}
	})

	t.Run("SetIf true", func(t *testing.T) {
		element := DIV().SetIf(true, "data-active", "yes")
		if element.attributes["data-active"] != "yes" {
			t.Error("Expected data-active='yes' when condition is true")
		}
	})

	t.Run("SetIf false", func(t *testing.T) {
		element := DIV().SetIf(false, "data-active", "yes")
		_, exists := element.attributes["data-active"]
		if exists {
			t.Error("Expected attribute not to be set when condition is false")
		}
	})

	t.Run("AddClassIf true", func(t *testing.T) {
		element := DIV().AddClassIf(true, "active")
		if !element.HasClass("active") {
			t.Error("Expected class 'active' when condition is true")
		}
	})

	t.Run("AddClassIf false", func(t *testing.T) {
		element := DIV().AddClassIf(false, "active")
		if element.HasClass("active") {
			t.Error("Expected no 'active' class when condition is false")
		}
	})
}

func TestChainedAttributes(t *testing.T) {
	// Test complex chaining of attributes
	element := DIV().
		ID("main-content").
		AddClass("container").
		AddClass("fluid").
		Style("margin: 20px;").
		Data("section", "main").
		Role("main").
		AriaLabel("Main content area").
		OnClick("handleMainClick()").
		TabIndex(0)

	// Verify all attributes are present
	expected := map[string]string{
		"id":            "main-content",
		"class":         "container fluid",
		"style":         "margin: 20px;",
		"data-section":  "main",
		"role":          "main",
		"aria-label":    "Main content area",
		"onclick":       "handleMainClick()",
		"tabindex":      "0",
	}

	for attr, expectedVal := range expected {
		actualVal := element.attributes[attr]
		if actualVal != expectedVal {
			t.Errorf("For attribute %q: expected %q, got %q", attr, expectedVal, actualVal)
		}
	}
}

func TestTableAttributes(t *testing.T) {
	cell := TD("Data").Colspan(2).Rowspan(3)
	
	if cell.attributes["colspan"] != "2" {
		t.Errorf("Expected colspan='2', got %q", cell.attributes["colspan"])
	}
	
	if cell.attributes["rowspan"] != "3" {
		t.Errorf("Expected rowspan='3', got %q", cell.attributes["rowspan"])
	}

	header := TH("Header").Scope("col")
	if header.attributes["scope"] != "col" {
		t.Errorf("Expected scope='col', got %q", header.attributes["scope"])
	}
}

func TestMediaAttributes(t *testing.T) {
	video := NewElement("video").
		Src("movie.mp4").
		Controls().
		AutoPlay().
		Loop().
		Muted().
		Preload("auto").
		Poster("poster.jpg").
		Width("640").
		Height("480")

	expected := map[string]string{
		"src":      "movie.mp4",
		"controls": "controls",
		"autoplay": "autoplay",
		"loop":     "loop",
		"muted":    "muted",
		"preload":  "auto",
		"poster":   "poster.jpg",
		"width":    "640",
		"height":   "480",
	}

	for attr, expectedVal := range expected {
		actualVal := video.attributes[attr]
		if actualVal != expectedVal {
			t.Errorf("For video attribute %q: expected %q, got %q", attr, expectedVal, actualVal)
		}
	}
}