package xssrender

// RenderGreeting builds a tiny HTML fragment that greets the caller.
//
// EXPERIMENT (intentionally unsafe): user-controlled name is concatenated
// into HTML without escaping. A secure implementation must use
// html.EscapeString (or equivalent) before embedding into markup.
func RenderGreeting(name string) string {
	return "<h1>Hello, " + name + "</h1>"
}
