package xssrender

// RenderGreeting builds a tiny HTML fragment that greets the caller.
// INTENTIONAL REGRESSION (Phase A validation): embed name without HTML escape
// so local /review-security can detect reflected XSS. Do not merge as-is.
func RenderGreeting(name string) string {
	return "<h1>Hello, " + name + "</h1>"
}
