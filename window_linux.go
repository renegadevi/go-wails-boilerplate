//go:build linux

package main

/*
#cgo pkg-config: gtk4

#include <gtk/gtk.h>
#include <stdlib.h>

static void paintFramelessWindow(const char *css) {
	GdkDisplay *display = gdk_display_get_default();
	if (display == NULL || css == NULL) {
		return;
	}
	GtkCssProvider *provider = gtk_css_provider_new();
	gtk_css_provider_load_from_string(provider, css);
	gtk_style_context_add_provider_for_display(
		display,
		GTK_STYLE_PROVIDER(provider),
		GTK_STYLE_PROVIDER_PRIORITY_USER);
	g_object_unref(provider);
}
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func init() {
	var once sync.Once
	hideLinuxFramelessEdge = func(window *application.WebviewWindow) {
		if !nativeFrameless || window == nil {
			return
		}
		// Light GTK themes (ZorinBlue-Light is #eaf0f6) paint the window
		// behind the webview. A frameless window leaves a strip of that
		// colour along the top. Paint the GTK window the page colour and
		// drop the theme's hairline and corner radius.
		css := fmt.Sprintf(`
window.wails-frameless {
  background-color: #%02x%02x%02x;
  border-radius: 0;
  box-shadow: 0 10px 30px 5px rgba(0, 0, 0, 0.28);
  border: none;
  outline: none;
  padding: 0;
}
window.wails-frameless #webview-box,
window.wails-frameless webkitwebview,
window.wails-frameless headerbar {
  background-color: #%02x%02x%02x;
  min-height: 0;
  margin: 0;
  padding: 0;
  border: none;
  box-shadow: none;
}
`, windowBackgroundR, windowBackgroundG, windowBackgroundB, windowBackgroundR, windowBackgroundG, windowBackgroundB)
		application.InvokeSync(func() {
			if window.NativeWindow() == nil {
				return
			}
			once.Do(func() {
				c := C.CString(css)
				defer C.free(unsafe.Pointer(c))
				C.paintFramelessWindow(c)
			})
		})
	}
}
