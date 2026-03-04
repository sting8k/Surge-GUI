//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

static NSStatusItem *surgeStatusItem = nil;
static NSImage *surgeTemplateIcon = nil;
static NSMenu *surgeStatusMenu = nil;

@interface SurgeStatusItemHandler : NSObject
- (void)openApp:(id)sender;
- (void)quitApp:(id)sender;
@end

@implementation SurgeStatusItemHandler
- (void)openApp:(id)sender {
    if ([NSApp isHidden]) {
        [NSApp unhide:nil];
    }
    [NSApp activateIgnoringOtherApps:YES];

    NSWindow *window = [NSApp keyWindow];
    if (window == nil && [[NSApp windows] count] > 0) {
        window = [[NSApp windows] objectAtIndex:0];
    }
    if (window != nil) {
        [window makeKeyAndOrderFront:nil];
    }
}

- (void)quitApp:(id)sender {
    [NSApp terminate:nil];
}
@end

static SurgeStatusItemHandler *surgeStatusHandler = nil;

@interface SurgeIconModeController : NSObject
+ (void)enableDockMode;
+ (void)enableMenuBarMode;
@end

@implementation SurgeIconModeController
+ (NSImage *)loadSurgeTemplateIcon {
    if (surgeTemplateIcon != nil) {
        return surgeTemplateIcon;
    }

    NSString *iconPath = [[NSBundle mainBundle] pathForResource:@"iconfile" ofType:@"icns"];
    if (iconPath == nil) {
        return nil;
    }

    NSImage *icon = [[NSImage alloc] initWithContentsOfFile:iconPath];
    if (icon == nil) {
        return nil;
    }

    [icon setSize:NSMakeSize(18, 18)];
    [icon setTemplate:YES];
    surgeTemplateIcon = icon;
    return surgeTemplateIcon;
}

+ (void)ensureMenu {
    if (surgeStatusMenu != nil) {
        return;
    }

    surgeStatusMenu = [[NSMenu alloc] initWithTitle:@"Surge"];
    if (surgeStatusHandler == nil) {
        surgeStatusHandler = [SurgeStatusItemHandler new];
    }

    NSMenuItem *openItem = [[NSMenuItem alloc] initWithTitle:@"Open Surge" action:@selector(openApp:) keyEquivalent:@""];
    [openItem setTarget:surgeStatusHandler];
    [surgeStatusMenu addItem:openItem];

    [surgeStatusMenu addItem:[NSMenuItem separatorItem]];

    NSMenuItem *quitItem = [[NSMenuItem alloc] initWithTitle:@"Quit Surge" action:@selector(quitApp:) keyEquivalent:@"q"];
    [quitItem setTarget:surgeStatusHandler];
    [surgeStatusMenu addItem:quitItem];
}

+ (void)ensureStatusItem {
    if (surgeStatusItem != nil) {
        return;
    }

    [self ensureMenu];

    surgeStatusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSSquareStatusItemLength];
    NSButton *button = [surgeStatusItem button];
    NSImage *icon = [self loadSurgeTemplateIcon];
    if (icon != nil) {
        [button setImage:icon];
        [button setImagePosition:NSImageOnly];
        [button setTitle:@""];
    } else {
        [button setTitle:@"⬇︎"];
    }

    [button setToolTip:@"Surge"];
    [surgeStatusItem setMenu:surgeStatusMenu];
}

+ (void)removeStatusItem {
    if (surgeStatusItem == nil) {
        return;
    }

    [[NSStatusBar systemStatusBar] removeStatusItem:surgeStatusItem];
    surgeStatusItem = nil;
}

+ (void)enableDockMode {
    [self removeStatusItem];
    [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];
    [NSApp activateIgnoringOtherApps:YES];
}

+ (void)enableMenuBarMode {
    [self ensureStatusItem];
    [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
}
@end

static void surge_set_icon_mode(int mode) {
    if (mode == 1) {
        [SurgeIconModeController performSelectorOnMainThread:@selector(enableMenuBarMode) withObject:nil waitUntilDone:YES];
        return;
    }
    [SurgeIconModeController performSelectorOnMainThread:@selector(enableDockMode) withObject:nil waitUntilDone:YES];
}
*/
import "C"

import "fmt"

func applyIconMode(mode string) error {
	switch mode {
	case iconModeDock:
		C.surge_set_icon_mode(0)
		return nil
	case iconModeMenuBar:
		C.surge_set_icon_mode(1)
		return nil
	default:
		return fmt.Errorf("unsupported icon mode: %s", mode)
	}
}
