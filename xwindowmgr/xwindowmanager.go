package xwindowmgr

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/keysym.h>
#include "x11gohelper.hpp"

#define MAX(a, b) ((a) > (b) ? (a) : (b))
*/
import "C"
import (
	"context"
	"fmt"
	"gopherwm/manager"
	"sync/atomic"
	"time"
)

type XWindowManager struct {
	xWindowContextManager manager.ObjectManager[xWindowContext]
	display               *C.Display
	rootID                uint64

	isClosing atomic.Bool
	loopDone  chan struct{}
}

func createXWindowContext(display *C.Display, window C.Window) xWindowContext {
	var ret xWindowContext
	ret.windowHandle = window

	var outWindow C.Window
	var outBorderWidth C.uint
	var outDepth C.uint

	C.XGetGeometry(display, ret.windowHandle, &outWindow, &(ret.rect.x), &(ret.rect.y), &(ret.rect.width), &(ret.rect.height), &outBorderWidth, &outDepth)
	fmt.Printf("CreatingContext.Window:%d\n", ret.windowHandle)
	return ret
}

func NewXWindowManager() (*XWindowManager, error) {
	tmpDisplay := C.XOpenDisplay(nil)
	if tmpDisplay == nil {
		return nil, fmt.Errorf("Cannot open display, res:%d")
	}

	ret := &XWindowManager{
		xWindowContextManager: *manager.NewObjectManager[xWindowContext](),
		display:               tmpDisplay,
		loopDone:              make(chan struct{}),
	}
	rootWindow := C._DefaultRootWindow(tmpDisplay)
	fmt.Printf("RootWindow::%d\n", rootWindow)
	rootWindowContext := createXWindowContext(tmpDisplay, rootWindow)
	ret.rootID = ret.xWindowContextManager.AddObject(rootWindowContext)
	return ret, nil
}

func (xm *XWindowManager) GrabKey(key uint64) int {
	rootCtx, err := xm.xWindowContextManager.GetObject(xm.rootID)
	if err != nil {
		fmt.Println(err.Error())
		return 0
	}
	generatedKey := C.XKeysymToKeycode(xm.display, C.ulong(key))
	ret := C.XGrabKey(
		xm.display,
		C.int(generatedKey),
		C.Mod1Mask,
		rootCtx.windowHandle,
		1,
		C.GrabModeAsync,
		C.GrabModeAsync,
	)

	return int(ret)
}

func (xm *XWindowManager) GrabButton(button uint32, mask uint32) int {
	rootCtx, err := xm.xWindowContextManager.GetObject(xm.rootID)
	if err != nil {
		fmt.Println(err.Error())
		return 0
	}
	ret := C.XGrabButton(
		xm.display,
		C.uint(button),
		C.Mod1Mask,
		rootCtx.windowHandle,
		1,
		C.uint(mask),
		C.GrabModeAsync,
		C.GrabModeAsync,
		C.None,
		C.None,
	)
	return int(ret)
}

func (xm *XWindowManager) SelectInput(inputMask int64) {
	rootCtx, err := xm.xWindowContextManager.GetObject(xm.rootID)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	C.XSelectInput(xm.display, rootCtx.windowHandle, C.long(inputMask))
}

func (xm *XWindowManager) Start(ctx context.Context) {

	go func() {
		defer close(xm.loopDone)

		fmt.Println("XWindowManager: event loop run")
		var ev C.XEvent
		var start C.XButtonEvent
		start.subwindow = C.None
		var attr C.XWindowAttributes
		for {
			for C.XPending(xm.display) != 0 {
				C.XNextEvent(xm.display, &ev)
				evType := unionToInt(ev)
				fmt.Println("XWindowManager:Event:", evType)
				if evType == C.KeyPress {
					fmt.Println("XWindowManager:Event:Key pressed:", unionToXKeyEvent(ev).keycode)
					if unionToXKeyEvent(ev).subwindow != C.None {
						C.XRaiseWindow(xm.display, unionToXKeyEvent(ev).subwindow)
					}
					// } else if evType == C.MapNotify {
					// 	fmt.Printf("Event: MapNotify\n")
					// 	addWindow(xm.display, &(unionToXMapEvent(ev).window), &windowSlice)
					// } else if evType == C.UnmapNotify {
					// 	fmt.Printf("Event: UnmapNotify\n")
					// 	removeWindow(&(unionToXMapEvent(ev).window), &windowSlice)
				} else if evType == C.ButtonPress {
					fmt.Println("XWindowManager:Event:Button pressed")
					if unionToXButtonEvent(ev).subwindow != C.None {
						C.XGetWindowAttributes(xm.display, unionToXButtonEvent(ev).subwindow, &attr)
						start = *unionToXButtonEvent(ev)
					}
				} else if evType == C.MotionNotify && start.subwindow != C.None {
					fmt.Println("XWindowManager:Event:Motion notify")
					xdiff := unionToXButtonEvent(ev).x_root - start.x_root
					ydiff := unionToXButtonEvent(ev).y_root - start.y_root

					var toDiffX C.int
					var toDiffY C.int

					if start.button == 1 {
						toDiffX = xdiff
						toDiffY = ydiff
					}

					var toWidth C.int
					var toHeight C.int

					if start.button == 3 {
						toWidth = xdiff
						toHeight = ydiff
					}

					C.XMoveResizeWindow(
						xm.display,
						start.subwindow,
						attr.x+toDiffX,
						attr.y+toDiffY,
						max(1, attr.width+toWidth),
						max(1, attr.height+toHeight))
				} else if evType == C.ButtonRelease {
					fmt.Println("XWindowManager:Event:Button release")
					start.subwindow = C.None
				}
			}

			if xm.isClosing.Load() {
				fmt.Println("XWindowManager: event loop is interrupted")
				return
			}
			time.Sleep(30 * time.Millisecond)
		}
	}()
	// for exit
	go func() {
		<-ctx.Done()
		fmt.Println("XWindowManager: context is canceled")
		xm.Close()
	}()
}

func (xm *XWindowManager) Close() {
	if xm.isClosing.Swap(true) {
		return
	}

	fmt.Println("Freeing XWindowManager resources")

	// wake up C.XNextEvent
	rootCtx, _ := xm.xWindowContextManager.GetObject(xm.rootID)
	C.send_quit_event(xm.display, rootCtx.windowHandle)

	// wait for goroutine with event loop is done
	<-xm.loopDone

	// closing display
	if xm.display != nil {
		C.XCloseDisplay(xm.display)
		xm.display = nil
		fmt.Println("XWindowManager resources freed")
	}
}
