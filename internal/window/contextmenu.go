package window

const (
	// ContextMenuDefault keeps the Wails behaviour. Release builds show the
	// webview menu only on inputs, editable content, and selected text. Dev
	// builds always show it, with Inspect Element.
	ContextMenuDefault = "default"
	// ContextMenuDisabled removes right-click everywhere, inputs included.
	// The frontend stops the event before any handler sees it.
	ContextMenuDisabled = "disabled"
	// ContextMenuCustom removes the webview menu but still delivers the event.
	// Use Vue @contextmenu handlers or a Wails --custom-contextmenu menu.
	ContextMenuCustom = "custom"
)

// contextMenu is the right-click behaviour. Pick one of the constants above.
// The dev-only inspector shortcut in internal/inspector works in every mode.
const contextMenu = ContextMenuDefault

// ContextMenu reports the right-click mode so the frontend can apply it.
func ContextMenu() string { return contextMenu }
