package mcpapps

import (
	_ "embed"
	"strings"
)

//go:embed app.html
var appHTMLTemplate string

//go:embed continuation.html
var continuationHTML string

// ContinuationHTML returns the dedicated Nexus-owned continuation controller.
// It is independent from node availability and from ordinary result cards.
func ContinuationHTML() string { return continuationHTML }

// HTML renders one shared MCP App document for a known AgentDock view.
// View and title are internal constants owned by AgentDock/NexusDock, not user input.
func HTML(view, title string) string {
	if view == "work_continuation" {
		return ContinuationHTML()
	}
	return strings.NewReplacer("{{VIEW}}", view, "{{TITLE}}", title).Replace(appHTMLTemplate)
}
