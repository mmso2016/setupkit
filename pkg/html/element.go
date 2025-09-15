// Package html provides programmatic HTML generation for SSR
// This package allows building HTML structures using Go functions like DIV(), A(), BR()
package html

import (
	"fmt"
	"html"
	"strings"
)

// Element represents an HTML element that can be rendered to string
type Element struct {
	tag        string
	attributes map[string]string
	children   []interface{} // Can contain Element, string, or other renderable content
	selfClose  bool          // For elements like <br>, <img>, etc.
}

// Attributes represents HTML attributes as key-value pairs
type Attributes map[string]string

// NewElement creates a new HTML element with the given tag
func NewElement(tag string) *Element {
	return &Element{
		tag:        tag,
		attributes: make(map[string]string),
		children:   make([]interface{}, 0),
		selfClose:  false,
	}
}

// NewSelfClosingElement creates a new self-closing HTML element (like br, img, hr)
func NewSelfClosingElement(tag string) *Element {
	return &Element{
		tag:        tag,
		attributes: make(map[string]string),
		children:   make([]interface{}, 0),
		selfClose:  true,
	}
}

// Attr sets an attribute on the element
func (e *Element) Attr(name, value string) *Element {
	e.attributes[name] = value
	return e
}

// Attrs sets multiple attributes on the element
func (e *Element) Attrs(attrs Attributes) *Element {
	for name, value := range attrs {
		e.attributes[name] = value
	}
	return e
}

// ID sets the id attribute
func (e *Element) ID(id string) *Element {
	return e.Attr("id", id)
}

// Class sets the class attribute
func (e *Element) Class(class string) *Element {
	return e.Attr("class", class)
}

// Style sets the style attribute
func (e *Element) Style(style string) *Element {
	return e.Attr("style", style)
}

// Text adds text content to the element (will be HTML escaped)
func (e *Element) Text(text string) *Element {
	e.children = append(e.children, html.EscapeString(text))
	return e
}

// HTML adds raw HTML content to the element (not escaped)
func (e *Element) HTML(content string) *Element {
	e.children = append(e.children, content)
	return e
}

// Child adds a child element
func (e *Element) Child(child *Element) *Element {
	e.children = append(e.children, child)
	return e
}

// Children adds multiple child elements
func (e *Element) Children(children ...*Element) *Element {
	for _, child := range children {
		e.children = append(e.children, child)
	}
	return e
}

// Render converts the element to HTML string
func (e *Element) Render() string {
	var result strings.Builder
	
	// Opening tag
	result.WriteString("<")
	result.WriteString(e.tag)
	
	// Attributes
	for name, value := range e.attributes {
		result.WriteString(fmt.Sprintf(` %s="%s"`, name, html.EscapeString(value)))
	}
	
	if e.selfClose {
		result.WriteString(" />")
		return result.String()
	}
	
	result.WriteString(">")
	
	// Children/content
	for _, child := range e.children {
		switch c := child.(type) {
		case *Element:
			result.WriteString(c.Render())
		case string:
			result.WriteString(c) // Assume already properly escaped if needed
		case fmt.Stringer:
			result.WriteString(c.String())
		default:
			result.WriteString(fmt.Sprintf("%v", c))
		}
	}
	
	// Closing tag
	result.WriteString("</")
	result.WriteString(e.tag)
	result.WriteString(">")
	
	return result.String()
}

// String implements Stringer interface
func (e *Element) String() string {
	return e.Render()
}