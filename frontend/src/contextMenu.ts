import { ContextMenu } from '../bindings/boilerplate/app'

// Capture on window runs before the Wails runtime listener and every element
// handler, so the mode holds in dev builds too, where Wails always shows the
// webview menu.
function blockContextMenu(event: MouseEvent): void {
  event.preventDefault()
  event.stopPropagation()
}

function suppressWebviewMenu(event: MouseEvent): void {
  event.preventDefault()
}

// applyContextMenuMode applies contextMenu from internal/window/contextmenu.go. The listener
// lives as long as the page.
export async function applyContextMenuMode(): Promise<void> {
  const mode = await ContextMenu()
  switch (mode) {
    case 'default':
      return
    case 'disabled':
      window.addEventListener('contextmenu', blockContextMenu, { capture: true })
      return
    case 'custom':
      window.addEventListener('contextmenu', suppressWebviewMenu, { capture: true })
      return
    default:
      throw new Error(`unknown context menu mode: ${mode}`)
  }
}
