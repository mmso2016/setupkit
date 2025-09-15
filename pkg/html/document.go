// Package html - Document builder for complete HTML pages
package html

// Document represents a complete HTML document
type Document struct {
	doctype string
	html    *Element
	head    *Element
	body    *Element
}

// NewDocument creates a new HTML document with standard structure
func NewDocument() *Document {
	head := HEAD()
	body := BODY()
	htmlElement := HTML().Children(head, body)
	
	return &Document{
		doctype: "<!DOCTYPE html>",
		html:    htmlElement,
		head:    head,
		body:    body,
	}
}

// SetTitle sets the document title
func (d *Document) SetTitle(title string) *Document {
	d.head.Child(TITLE(title))
	return d
}

// SetCharset sets the character encoding (default: utf-8)
func (d *Document) SetCharset(charset string) *Document {
	d.head.Child(META().Attr("charset", charset))
	return d
}

// SetViewport sets the viewport meta tag for responsive design
func (d *Document) SetViewport(content string) *Document {
	if content == "" {
		content = "width=device-width, initial-scale=1.0"
	}
	d.head.Child(META().Attr("name", "viewport").Attr("content", content))
	return d
}

// AddCSS adds CSS styles to the document
func (d *Document) AddCSS(css string) *Document {
	d.head.Child(STYLE(css))
	return d
}

// AddExternalCSS links to an external CSS file
func (d *Document) AddExternalCSS(href string) *Document {
	d.head.Child(LINK().Attr("rel", "stylesheet").Attr("href", href))
	return d
}

// AddJS adds JavaScript to the document
func (d *Document) AddJS(js string) *Document {
	d.head.Child(SCRIPT(js))
	return d
}

// AddExternalJS links to an external JavaScript file
func (d *Document) AddExternalJS(src string) *Document {
	d.head.Child(SCRIPT("").Attr("src", src))
	return d
}

// AddMeta adds a meta tag to the head
func (d *Document) AddMeta(name, content string) *Document {
	d.head.Child(META().Attr("name", name).Attr("content", content))
	return d
}

// AddToHead adds any element to the document head
func (d *Document) AddToHead(element *Element) *Document {
	d.head.Child(element)
	return d
}

// AddToBody adds any element to the document body
func (d *Document) AddToBody(element *Element) *Document {
	d.body.Child(element)
	return d
}

// SetBodyContent sets the entire body content (replaces existing)
func (d *Document) SetBodyContent(elements ...*Element) *Document {
	d.body = BODY().Children(elements...)
	d.html.children[1] = d.body // Replace body in html element
	return d
}

// GetHead returns the head element for direct manipulation
func (d *Document) GetHead() *Element {
	return d.head
}

// GetBody returns the body element for direct manipulation
func (d *Document) GetBody() *Element {
	return d.body
}

// Render renders the complete HTML document
func (d *Document) Render() string {
	return d.doctype + "\n" + d.html.Render()
}

// String implements Stringer interface
func (d *Document) String() string {
	return d.Render()
}

// Common CSS framework and utility functions

// AddBootstrapCSS adds Bootstrap CSS CDN link
func (d *Document) AddBootstrapCSS() *Document {
	return d.AddExternalCSS("https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css")
}

// AddBootstrapJS adds Bootstrap JavaScript CDN link
func (d *Document) AddBootstrapJS() *Document {
	return d.AddExternalJS("https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js")
}

// AddTailwindCSS adds Tailwind CSS CDN link
func (d *Document) AddTailwindCSS() *Document {
	return d.AddExternalCSS("https://cdn.tailwindcss.com")
}

// AddDefaultSetupKitStyles adds default styles for SetupKit installers
func (d *Document) AddDefaultSetupKitStyles() *Document {
	css := `
		body {
			font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
			margin: 0;
			padding: 20px;
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			min-height: 100vh;
			color: white;
		}
		.container {
			max-width: 800px;
			margin: 0 auto;
			background: rgba(255,255,255,0.1);
			border-radius: 15px;
			padding: 40px;
			backdrop-filter: blur(10px);
			box-shadow: 0 8px 32px rgba(0,0,0,0.1);
		}
		.header {
			text-align: center;
			margin-bottom: 40px;
		}
		.title {
			font-size: 2.5rem;
			font-weight: 300;
			margin-bottom: 10px;
			text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
		}
		.subtitle {
			font-size: 1.1rem;
			opacity: 0.9;
			margin-bottom: 5px;
		}
		.version {
			font-size: 0.9rem;
			opacity: 0.7;
		}
		.component {
			background: rgba(255,255,255,0.1);
			margin: 15px 0;
			padding: 20px;
			border-radius: 10px;
			border-left: 4px solid #4CAF50;
			transition: all 0.3s ease;
		}
		.component:hover {
			background: rgba(255,255,255,0.15);
			transform: translateY(-2px);
		}
		.component.required {
			border-left-color: #ff9800;
		}
		.component-header {
			display: flex;
			align-items: center;
			margin-bottom: 10px;
		}
		.component-name {
			font-size: 1.2rem;
			font-weight: 600;
			flex-grow: 1;
		}
		.component-size {
			font-size: 0.9rem;
			opacity: 0.8;
		}
		.component-description {
			opacity: 0.9;
			line-height: 1.4;
		}
		.button {
			background: rgba(255,255,255,0.2);
			border: 2px solid rgba(255,255,255,0.3);
			color: white;
			padding: 12px 30px;
			border-radius: 25px;
			font-size: 1rem;
			font-weight: 600;
			cursor: pointer;
			transition: all 0.3s ease;
			margin: 10px;
		}
		.button:hover {
			background: rgba(255,255,255,0.3);
			border-color: rgba(255,255,255,0.5);
			transform: translateY(-2px);
		}
		.button.primary {
			background: #4CAF50;
			border-color: #4CAF50;
		}
		.button.primary:hover {
			background: #45a049;
			border-color: #45a049;
		}
		.progress {
			width: 100%;
			height: 20px;
			background: rgba(255,255,255,0.2);
			border-radius: 10px;
			overflow: hidden;
			margin: 20px 0;
		}
		.progress-bar {
			height: 100%;
			background: linear-gradient(90deg, #4CAF50, #45a049);
			transition: width 0.3s ease;
		}
	`
	return d.AddCSS(css)
}