//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework QuartzCore

#import <Cocoa/Cocoa.h>
#import <QuartzCore/QuartzCore.h>
#import <objc/runtime.h>
#import <math.h>

// Keep these in sync with .titlebar--mac and .titlebar-lights in
// frontend/src/style.css. The bar is the taller custom toolbar used by
// current macOS apps. The lights are 14pt, so a 42pt bar centers them
// 14pt from the top. They sit 16pt in from the left, clear of the corner.
static const CGFloat kMacToolbarHeight = 42;
static const CGFloat kMacLightInset = 16;

// Wails already steps a normal zoom so the page resizes with the window.
// While it measures the target, animationResizeTime: returns 0. A second
// animator treats that 0 as "too short" and runs the zoom at its 0.12s
// floor, about twice the real duration. Leave zoom to Wails.

static BOOL macPositioning = NO;
static BOOL macObserverInstalled = NO;
static BOOL macEnteringFullScreen = NO;
static BOOL macExitingFullScreen = NO;
static int macInstallAttempts = 0;

static void setFrameNow(NSView *view, NSRect frame) {
	[CATransaction begin];
	[CATransaction setDisableActions:YES];
	[view.layer removeAllAnimations];
	[NSAnimationContext beginGrouping];
	[[NSAnimationContext currentContext] setDuration:0];
	[[NSAnimationContext currentContext] setAllowsImplicitAnimation:NO];
	[view setFrame:frame];
	[NSAnimationContext endGrouping];
	[CATransaction commit];
}

static NSWindow *macFramelessWindow(void) {
	for (NSWindow *window in [NSApp windows]) {
		if (window.titlebarAppearsTransparent) {
			return window;
		}
	}
	return nil;
}

static void placeMacTrafficLights(NSWindow *window) {
	if (window == nil || macPositioning || macEnteringFullScreen) {
		return;
	}
	// Stable fullscreen hides the lights. While leaving fullscreen the mask
	// stays set for the whole animation, and AppKit slides the lights back to
	// the short system title bar. Keep placing through that animation so they
	// come back centered instead of jumping down at the end.
	BOOL fullscreen = (window.styleMask & NSWindowStyleMaskFullScreen) == NSWindowStyleMaskFullScreen;
	if (fullscreen && !macExitingFullScreen) {
		return;
	}

	NSButton *close = [window standardWindowButton:NSWindowCloseButton];
	NSButton *mini = [window standardWindowButton:NSWindowMiniaturizeButton];
	NSButton *zoom = [window standardWindowButton:NSWindowZoomButton];
	if (close == nil || mini == nil || zoom == nil || close.frame.size.height < 1) {
		return;
	}

	// close -> NSTitlebarView -> NSTitlebarContainerView -> NSThemeFrame
	NSView *titlebar = close.superview;
	NSView *container = titlebar.superview;
	NSView *theme = container.superview;
	if (titlebar == nil || container == nil || theme == nil) {
		return;
	}

	CGFloat pitch = mini.frame.origin.x - close.frame.origin.x;
	if (pitch < 1) {
		pitch = 23;
	}
	CGFloat y = (kMacToolbarHeight - close.frame.size.height) / 2.0;
	CGFloat themeHeight = theme.frame.size.height;
	BOOL placed = fabs(container.frame.size.height - kMacToolbarHeight) < 0.5 &&
		fabs(container.frame.origin.y - (themeHeight - kMacToolbarHeight)) < 0.5 &&
		fabs(close.frame.origin.y - y) < 0.5 &&
		fabs(close.frame.origin.x - kMacLightInset) < 0.5;
	if (placed && !macExitingFullScreen) {
		return;
	}

	macPositioning = YES;

	NSRect containerFrame = container.frame;
	containerFrame.origin.x = 0;
	containerFrame.origin.y = themeHeight - kMacToolbarHeight;
	containerFrame.size.width = theme.frame.size.width;
	containerFrame.size.height = kMacToolbarHeight;
	setFrameNow(container, containerFrame);

	NSRect titlebarFrame = titlebar.frame;
	titlebarFrame.origin = NSMakePoint(0, 0);
	titlebarFrame.size = NSMakeSize(containerFrame.size.width, kMacToolbarHeight);
	setFrameNow(titlebar, titlebarFrame);

	NSButton *buttons[3] = {close, mini, zoom};
	for (int i = 0; i < 3; i++) {
		NSRect frame = buttons[i].frame;
		frame.origin.x = kMacLightInset + (i * pitch);
		frame.origin.y = y;
		setFrameNow(buttons[i], frame);
	}

	macPositioning = NO;
}

// Leaving fullscreen lays the title bar out for the whole exit animation.
// That layout runs after DidResize and puts the lights back on the short
// system bar, which is the jump once the animation ends. Placing at the
// end of layout keeps them in the toolbar while the window is moving.
static IMP macOrigTitlebarLayout = NULL;
static IMP macOrigContainerLayout = NULL;

static void macLayoutAndPlace(id self, IMP orig) {
	((void (*)(id, SEL))orig)(self, @selector(layout));
	if (macPositioning) {
		return;
	}
	NSWindow *window = [self window];
	if (window == nil || !window.titlebarAppearsTransparent) {
		return;
	}
	placeMacTrafficLights(window);
}

static void macTitlebarLayout(id self, SEL cmd) {
	(void)cmd;
	macLayoutAndPlace(self, macOrigTitlebarLayout);
}

static void macContainerLayout(id self, SEL cmd) {
	(void)cmd;
	macLayoutAndPlace(self, macOrigContainerLayout);
}

static BOOL macTitlebarLayoutInstalled = NO;

static void macInstallTitlebarLayout(void) {
	if (macTitlebarLayoutInstalled) {
		return;
	}
	macTitlebarLayoutInstalled = YES;

	Class titlebar = NSClassFromString(@"NSTitlebarView");
	Class container = NSClassFromString(@"NSTitlebarContainerView");
	Method titlebarLayout = class_getInstanceMethod(titlebar, @selector(layout));
	Method containerLayout = class_getInstanceMethod(container, @selector(layout));
	if (titlebarLayout != NULL) {
		macOrigTitlebarLayout = method_getImplementation(titlebarLayout);
		method_setImplementation(titlebarLayout, (IMP)macTitlebarLayout);
	}
	if (containerLayout != NULL) {
		macOrigContainerLayout = method_getImplementation(containerLayout);
		method_setImplementation(containerLayout, (IMP)macContainerLayout);
	}
}

static void installMacTrafficLights(void) {
	dispatch_async(dispatch_get_main_queue(), ^{
		NSWindow *window = macFramelessWindow();
		if (window == nil) {
			if (macInstallAttempts >= 30) {
				return;
			}
			macInstallAttempts++;
			dispatch_after(dispatch_time(DISPATCH_TIME_NOW, (int64_t)(0.1 * NSEC_PER_SEC)), dispatch_get_main_queue(), ^{
				installMacTrafficLights();
			});
			return;
		}

		placeMacTrafficLights(window);
		if (macObserverInstalled) {
			return;
		}
		macObserverInstalled = YES;
		macInstallTitlebarLayout();

		NSNotificationCenter *center = [NSNotificationCenter defaultCenter];
		NSOperationQueue *main = [NSOperationQueue mainQueue];
		void (^place)(NSNotification *) = ^(NSNotification *note) {
			placeMacTrafficLights((NSWindow *)note.object);
		};
		[center addObserverForName:NSWindowDidResizeNotification object:window queue:main usingBlock:place];
		[center addObserverForName:NSWindowDidBecomeKeyNotification object:window queue:main usingBlock:place];
		[center addObserverForName:NSWindowWillEnterFullScreenNotification object:window queue:main usingBlock:^(NSNotification *note) {
			(void)note;
			macEnteringFullScreen = YES;
			macExitingFullScreen = NO;
			// DidEnter clears this. If the transition is cancelled, no notification
			// follows, so drop the guard after the animation would have finished.
			dispatch_after(dispatch_time(DISPATCH_TIME_NOW, (int64_t)(1.5 * NSEC_PER_SEC)), dispatch_get_main_queue(), ^{
				if (!macEnteringFullScreen) {
					return;
				}
				macEnteringFullScreen = NO;
				placeMacTrafficLights(window);
			});
		}];
		[center addObserverForName:NSWindowDidEnterFullScreenNotification object:window queue:main usingBlock:^(NSNotification *note) {
			(void)note;
			macEnteringFullScreen = NO;
		}];
		[center addObserverForName:NSWindowWillExitFullScreenNotification object:window queue:main usingBlock:^(NSNotification *note) {
			(void)note;
			macEnteringFullScreen = NO;
			macExitingFullScreen = YES;
			placeMacTrafficLights(window);
		}];
		[center addObserverForName:NSWindowDidExitFullScreenNotification object:window queue:main usingBlock:^(NSNotification *note) {
			(void)note;
			macExitingFullScreen = NO;
			placeMacTrafficLights(window);
		}];

		// AppKit lays the standard title bar out again when the window is shown.
		dispatch_after(dispatch_time(DISPATCH_TIME_NOW, (int64_t)(0.2 * NSEC_PER_SEC)), dispatch_get_main_queue(), ^{
			placeMacTrafficLights(window);
		});
	});
}
*/
import "C"

func init() {
	installMacTrafficLights = func() {
		C.installMacTrafficLights()
	}
}
