package xssrender

import "html"

// RenderGreeting builds a tiny HTML fragment that greets the caller.
// User-controlled name is HTML-escaped before embedding into markup
// to prevent reflected XSS.
func RenderGreeting(name string) string {
	return "<h1>Hello, " + html.EscapeString(name) + "</h1>"
}
