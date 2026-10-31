//go:build darwin && (!production || devtools)

package inspector

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework WebKit

#import <Cocoa/Cocoa.h>
#import <WebKit/WebKit.h>

// Dev builds only. These are private WebKit calls, so they must never reach
// a release or App Store build. Wails sets the public webView.inspectable,
// which only exposes the page to Safari's Develop menu.

static WKWebView *macWebView(void *window) {
	NSWindow *nsWindow = (NSWindow *)window;
	if (nsWindow == nil || ![nsWindow respondsToSelector:@selector(webView)]) {
		return nil;
	}
	id view = [nsWindow performSelector:@selector(webView)];
	if (![view isKindOfClass:[WKWebView class]]) {
		return nil;
	}
	return (WKWebView *)view;
}

static void enableMacInspectElement(void *window) {
	WKWebView *view = macWebView(window);
	if (view == nil) {
		return;
	}
	@try {
		[view.configuration.preferences setValue:@YES forKey:@"developerExtrasEnabled"];
	} @catch (NSException *exception) {
		NSLog(@"Could not enable Inspect Element: %@", exception.reason);
	}
}

static void openMacInspector(void *window) {
	dispatch_async(dispatch_get_main_queue(), ^{
		WKWebView *view = macWebView(window);
		if (view == nil || ![view respondsToSelector:@selector(_inspector)]) {
			return;
		}
		@try {
			id inspector = [view performSelector:@selector(_inspector)];
			if ([inspector respondsToSelector:@selector(show)]) {
				[inspector performSelector:@selector(show)];
			}
		} @catch (NSException *exception) {
			NSLog(@"Could not open the inspector: %@", exception.reason);
		}
	});
}
*/
import "C"

import "github.com/wailsapp/wails/v3/pkg/application"

func init() {
	enableInspectElement = func(window *application.WebviewWindow) {
		if window == nil {
			return
		}
		application.InvokeSync(func() {
			if nativeWindow := window.NativeWindow(); nativeWindow != nil {
				C.enableMacInspectElement(nativeWindow)
			}
		})
	}
	openInspector = func(window application.Window) {
		if nativeWindow := window.NativeWindow(); nativeWindow != nil {
			C.openMacInspector(nativeWindow)
		}
	}
}
