package html

import (
	"strings"
	"testing"
)

func TestBasicElements(t *testing.T) {
	tests := []struct {
		name     string
		element  *Element
		expected string
	}{
		{
			name:     "Simple DIV",
			element:  DIV(),
			expected: "<div></div>",
		},
		{
			name:     "DIV with text",
			element:  DIV().Text("Hello World"),
			expected: "<div>Hello World</div>",
		},
		{
			name:     "DIV with ID",
			element:  DIV().ID("myDiv"),
			expected: `<div id="myDiv"></div>`,
		},
		{
			name:     "DIV with class",
			element:  DIV().Class("container"),
			expected: `<div class="container"></div>`,
		},
		{
			name:     "P with text",
			element:  P("This is a paragraph."),
			expected: "<p>This is a paragraph.</p>",
		},
		{
			name:     "H1 with text",
			element:  H1("Main Title"),
			expected: "<h1>Main Title</h1>",
		},
		{
			name:     "A with href",
			element:  A("https://example.com", "Click here"),
			expected: `<a href="https://example.com">Click here</a>`,
		},
		{
			name:     "Self-closing BR",
			element:  BR(),
			expected: "<br />",
		},
		{
			name:     "Self-closing HR",
			element:  HR(),
			expected: "<hr />",
		},
		{
			name:     "IMG with src and alt",
			element:  IMG("image.jpg", "Description"),
			expected: `<img src="image.jpg" alt="Description" />`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.element.Render()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestNestedElements(t *testing.T) {
	// Create a nested structure: <div class="container"><h1>Title</h1><p>Content</p></div>
	container := DIV().Class("container").Children(
		H1("Title"),
		P("Content"),
	)

	expected := `<div class="container"><h1>Title</h1><p>Content</p></div>`
	result := container.Render()

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestAttributes(t *testing.T) {
	element := DIV().Attrs(Attributes{
		"id":    "test",
		"class": "container active",
		"data-toggle": "modal",
	})

	result := element.Render()
	
	// Check that all attributes are present (order may vary)
	expectedAttrs := []string{`id="test"`, `class="container active"`, `data-toggle="modal"`}
	for _, attr := range expectedAttrs {
		if !strings.Contains(result, attr) {
			t.Errorf("Expected result to contain %q, got %q", attr, result)
		}
	}
}

func TestHTMLEscaping(t *testing.T) {
	// Test that text content is properly escaped
	element := DIV().Text("<script>alert('xss')</script>")
	result := element.Render()
	
	expected := "<div>&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;</div>"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestRawHTML(t *testing.T) {
	// Test that HTML content is not escaped
	element := DIV().HTML("<strong>Bold text</strong>")
	result := element.Render()
	
	expected := "<div><strong>Bold text</strong></div>"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestComplexStructure(t *testing.T) {
	// Build a complex HTML structure
	page := DIV().Class("page").Children(
		HEADER().Child(
			H1("Welcome to SetupKit"),
		),
		MAIN().Class("content").Children(
			SECTION().Children(
				H2("Components"),
				UL().Children(
					LI("Core Application"),
					LI("Documentation"),
					LI("Examples"),
				),
			),
		),
		FOOTER().Child(
			P("© 2025 SetupKit Framework"),
		),
	)

	result := page.Render()
	
	// Check for key parts of the structure
	expectedParts := []string{
		`<div class="page">`,
		"<header>",
		"<h1>Welcome to SetupKit</h1>",
		`<main class="content">`,
		"<section>",
		"<h2>Components</h2>",
		"<ul>",
		"<li>Core Application</li>",
		"<li>Documentation</li>",
		"<li>Examples</li>",
		"</ul>",
		"</section>",
		"</main>",
		"<footer>",
		"<p>© 2025 SetupKit Framework</p>",
		"</footer>",
		"</div>",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected result to contain %q", part)
		}
	}
}

func TestDocument(t *testing.T) {
	doc := NewDocument().
		SetTitle("Test Page").
		SetCharset("utf-8").
		SetViewport("").
		AddCSS("body { margin: 0; }").
		AddToBody(
			DIV().Class("container").Child(
				H1("Hello World"),
			),
		)

	result := doc.Render()

	expectedParts := []string{
		"<!DOCTYPE html>",
		"<html>",
		"<head>",
		"<title>Test Page</title>",
		`<meta charset="utf-8" />`,
		`<meta name="viewport" content="width=device-width, initial-scale=1.0" />`,
		"<style>body { margin: 0; }</style>",
		"</head>",
		"<body>",
		`<div class="container">`,
		"<h1>Hello World</h1>",
		"</div>",
		"</body>",
		"</html>",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected result to contain %q", part)
		}
	}
}