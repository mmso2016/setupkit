// Package html - Standard HTML element functions
// This file provides convenient functions for creating common HTML elements
package html

// Document structure elements

// HTML creates an <html> element
func HTML() *Element {
	return NewElement("html")
}

// HEAD creates a <head> element
func HEAD() *Element {
	return NewElement("head")
}

// BODY creates a <body> element
func BODY() *Element {
	return NewElement("body")
}

// TITLE creates a <title> element
func TITLE(text string) *Element {
	return NewElement("title").Text(text)
}

// META creates a <meta> element
func META() *Element {
	return NewSelfClosingElement("meta")
}

// LINK creates a <link> element
func LINK() *Element {
	return NewSelfClosingElement("link")
}

// STYLE creates a <style> element
func STYLE(css string) *Element {
	return NewElement("style").HTML(css)
}

// SCRIPT creates a <script> element
func SCRIPT(js string) *Element {
	return NewElement("script").HTML(js)
}

// Block elements

// DIV creates a <div> element
func DIV() *Element {
	return NewElement("div")
}

// P creates a <p> element
func P(text ...string) *Element {
	p := NewElement("p")
	for _, t := range text {
		p.Text(t)
	}
	return p
}

// PRE creates a <pre> element
func PRE(text string) *Element {
	return NewElement("pre").Text(text)
}

// H1 creates an <h1> element
func H1(text string) *Element {
	return NewElement("h1").Text(text)
}

// H2 creates an <h2> element
func H2(text string) *Element {
	return NewElement("h2").Text(text)
}

// H3 creates an <h3> element
func H3(text string) *Element {
	return NewElement("h3").Text(text)
}

// H4 creates an <h4> element
func H4(text string) *Element {
	return NewElement("h4").Text(text)
}

// H5 creates an <h5> element
func H5(text string) *Element {
	return NewElement("h5").Text(text)
}

// H6 creates an <h6> element
func H6(text string) *Element {
	return NewElement("h6").Text(text)
}

// SECTION creates a <section> element
func SECTION() *Element {
	return NewElement("section")
}

// HEADER creates a <header> element
func HEADER() *Element {
	return NewElement("header")
}

// FOOTER creates a <footer> element
func FOOTER() *Element {
	return NewElement("footer")
}

// MAIN creates a <main> element
func MAIN() *Element {
	return NewElement("main")
}

// ARTICLE creates an <article> element
func ARTICLE() *Element {
	return NewElement("article")
}

// ASIDE creates an <aside> element
func ASIDE() *Element {
	return NewElement("aside")
}

// NAV creates a <nav> element
func NAV() *Element {
	return NewElement("nav")
}

// Inline elements

// A creates an <a> element (anchor/link)
func A(href string, text ...string) *Element {
	a := NewElement("a").Attr("href", href)
	for _, t := range text {
		a.Text(t)
	}
	return a
}

// SPAN creates a <span> element
func SPAN(text ...string) *Element {
	span := NewElement("span")
	for _, t := range text {
		span.Text(t)
	}
	return span
}

// STRONG creates a <strong> element
func STRONG(text string) *Element {
	return NewElement("strong").Text(text)
}

// EM creates an <em> element
func EM(text string) *Element {
	return NewElement("em").Text(text)
}

// B creates a <b> element
func B(text string) *Element {
	return NewElement("b").Text(text)
}

// I creates an <i> element
func I(text string) *Element {
	return NewElement("i").Text(text)
}

// CODE creates a <code> element
func CODE(text string) *Element {
	return NewElement("code").Text(text)
}

// Self-closing elements

// BR creates a <br> element (line break)
func BR() *Element {
	return NewSelfClosingElement("br")
}

// HR creates an <hr> element (horizontal rule)
func HR() *Element {
	return NewSelfClosingElement("hr")
}

// IMG creates an <img> element
func IMG(src, alt string) *Element {
	return NewSelfClosingElement("img").Attr("src", src).Attr("alt", alt)
}

// INPUT creates an <input> element
func INPUT(inputType string) *Element {
	return NewSelfClosingElement("input").Attr("type", inputType)
}

// Lists

// UL creates a <ul> element (unordered list)
func UL() *Element {
	return NewElement("ul")
}

// OL creates an <ol> element (ordered list)
func OL() *Element {
	return NewElement("ol")
}

// LI creates an <li> element (list item)
func LI(text ...string) *Element {
	li := NewElement("li")
	for _, t := range text {
		li.Text(t)
	}
	return li
}

// DL creates a <dl> element (description list)
func DL() *Element {
	return NewElement("dl")
}

// DT creates a <dt> element (description term)
func DT(text string) *Element {
	return NewElement("dt").Text(text)
}

// DD creates a <dd> element (description definition)
func DD(text string) *Element {
	return NewElement("dd").Text(text)
}

// Tables

// TABLE creates a <table> element
func TABLE() *Element {
	return NewElement("table")
}

// THEAD creates a <thead> element
func THEAD() *Element {
	return NewElement("thead")
}

// TBODY creates a <tbody> element
func TBODY() *Element {
	return NewElement("tbody")
}

// TFOOT creates a <tfoot> element
func TFOOT() *Element {
	return NewElement("tfoot")
}

// TR creates a <tr> element (table row)
func TR() *Element {
	return NewElement("tr")
}

// TH creates a <th> element (table header cell)
func TH(text string) *Element {
	return NewElement("th").Text(text)
}

// TD creates a <td> element (table data cell)
func TD(text string) *Element {
	return NewElement("td").Text(text)
}

// Forms

// FORM creates a <form> element
func FORM() *Element {
	return NewElement("form")
}

// LABEL creates a <label> element
func LABEL(text string) *Element {
	return NewElement("label").Text(text)
}

// BUTTON creates a <button> element
func BUTTON(text string) *Element {
	return NewElement("button").Text(text)
}

// TEXTAREA creates a <textarea> element
func TEXTAREA(text ...string) *Element {
	textarea := NewElement("textarea")
	for _, t := range text {
		textarea.Text(t)
	}
	return textarea
}

// SELECT creates a <select> element
func SELECT() *Element {
	return NewElement("select")
}

// OPTION creates an <option> element
func OPTION(value, text string) *Element {
	return NewElement("option").Attr("value", value).Text(text)
}

// FIELDSET creates a <fieldset> element
func FIELDSET() *Element {
	return NewElement("fieldset")
}

// LEGEND creates a <legend> element
func LEGEND(text string) *Element {
	return NewElement("legend").Text(text)
}