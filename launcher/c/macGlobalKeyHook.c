// alterkeys.c
// http://osxbook.com
//
// Complile using the following command line:
//     gcc -Wall -o alterkeys alterkeys.c -framework ApplicationServices
//
// You need superuser privileges to create the event tap, unless accessibility
// is enabled. To do so, select the "Enable access for assistive devices"
// checkbox in the Universal Access system preference pane.

#include <ApplicationServices/ApplicationServices.h>

uint8_t HandleKey(int,int);
int menu_active = 0;
 CFMachPortRef      eventTap;
// This callback will be invoked every time there is a keystroke.
//
CGEventRef
myCGEventCallback(CGEventTapProxy proxy, CGEventType type,
                  CGEventRef event, void *refcon)
{
    // Paranoid sanity check.
    //if ((type != kCGEventKeyDown) && (type != kCGEventKeyUp))
    if ((type != kCGEventKeyUp) && (type != kCGEventKeyDown))
        return event;
    
    // The incoming keycode.
    CGKeyCode keycode = (CGKeyCode)CGEventGetIntegerValueField( event, kCGKeyboardEventKeycode);

    printf("C callback Key: %i\n  Calling HandleKey\n  Keydown:%s\n", keycode,type == kCGEventKeyDown? "yes" : "no" );
    
    uint8_t ret = HandleKey((int) keycode, type == kCGEventKeyDown? 1 : 0);
    if (ret>0) {
        printf("  Golang used key, returning NULL to event tap\n");
        return NULL;
    }

    printf("  Golang did not use key, returning event to event tap\n");
    
    // Set the modified keycode field in the event.
    //CGEventSetIntegerValueField( event, kCGKeyboardEventKeycode, (int64_t)keycode);
    
    // We must return the event for it to be useful.
    return event;
}

int
watchKeys(void)
{
   
    CGEventMask        eventMask;
    CFRunLoopSourceRef runLoopSource;
    
    // Create an event tap. We are interested in key presses.
    eventMask = ((1 << kCGEventKeyDown) | (1 << kCGEventKeyUp));
    eventTap = CGEventTapCreate(kCGSessionEventTap, kCGHeadInsertEventTap, 0,
                                eventMask, myCGEventCallback, NULL);
    if (!eventTap) {
        fprintf(stderr, "failed to create event tap\n");
        exit(1);
    }
    
    // Create a run loop source.
    runLoopSource = CFMachPortCreateRunLoopSource( kCFAllocatorDefault, eventTap, 0);
    
    // Add to the current run loop.
    CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
    
    // Enable the event tap.
    CGEventTapEnable(eventTap, true);
    
    // Set it all running.
    CFRunLoopRun();
    
    // In a real program, one would have arranged for cleaning up.
    
    exit(0);
}

void ReEnableEventTap() {
    CGEventTapEnable(eventTap, true);
}

void DisableEventTap() {
    CGEventTapEnable(eventTap, false);
}