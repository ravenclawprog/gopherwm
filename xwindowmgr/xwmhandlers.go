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
	"fmt"
	"unsafe"
)

func (xm *XWindowManager) handleMapRequest(evWindow C.Window, data *C.XMapRequestEvent) {
	fmt.Println("XWM: Handling MapRequest for window \"", evWindow, "\"")

	clientContext, _ := xm.clientManager.FindByClient(evWindow)

	if clientContext != nil && clientContext.Frame != C.None {
		C.XMapWindow(xm.backend.display, clientContext.Frame)
		C.XMapWindow(xm.backend.display, clientContext.Window)
	} else {
		newClient := generateClientContext(xm.backend, evWindow)
		newClient.CreateFrame(xm.backend, xm.rootWindow)
		fmt.Println("XWM: Handling MapRequest for window \"", evWindow, "\"; create frame window \"", newClient.Frame, "\"; Window Title:", newClient.Title)
		xm.clientManager.Add(newClient)
		C.XMapWindow(xm.backend.display, newClient.Frame)
		C.XMapWindow(xm.backend.display, newClient.Window)
	}
	xm.backend.Flush()
}

func (xm *XWindowManager) handleUnmapNotify(evWindow C.Window, data *C.XUnmapEvent) {
	fmt.Println("XWM: Handling UnmapNotify for window \"", evWindow, "\"")

	var targetFrame C.Window = C.None
	clientContext, err := xm.clientManager.FindByClient(evWindow)
	if err != nil {
		fmt.Println("XWM: Handling UnmapNotify: cannot find clientWindow:", err.Error())
		return
	}
	targetFrame = clientContext.Frame
	if targetFrame != C.None {
		C.XUnmapSubwindows(xm.backend.display, clientContext.Frame)
		C.XUnmapWindow(xm.backend.display, clientContext.Frame)
		xm.backend.Flush()
	}
}

func (xm *XWindowManager) handleDestroyNotify(evWindow C.Window, data *C.XDestroyWindowEvent) {
	fmt.Println("XWM: Handling Destroy Notify for window \"", evWindow, "\"")
	var targetFrame C.Window = C.None
	clientContext, err := xm.clientManager.FindByClient(evWindow)
	if err != nil {
		fmt.Println("XWM: Handling Destroy Notify: cannot find clientWindow:", err.Error())
		return
	}
	targetFrame = clientContext.Frame
	if targetFrame != C.None {
		C.XUnmapSubwindows(xm.backend.display, targetFrame)
		C.XUnmapWindow(xm.backend.display, targetFrame)
		C.XDestroyWindow(xm.backend.display, targetFrame)
	}
	xm.clientManager.Remove(clientContext.Window)
	xm.backend.Flush()
}

func (xm *XWindowManager) handleKeyPress(evWindow C.Window, data *C.XKeyEvent) {
	fmt.Println("XWM: Handling KeyPress for window \"", evWindow, "\"")

	if data.subwindow != C.None {
		C.XRaiseWindow(xm.backend.display, data.subwindow)
	}
	cleanModifiers := data.state &^ C.uint(C.Mod2Mask|C.LockMask)
	targetCode := C.XKeysymToKeycode(xm.backend.display, C.KeySym(C.XK_Return))
	if data.keycode == C.uint(targetCode) && cleanModifiers == C.uint(C.Mod4Mask) {
		fmt.Println("Win + Enter! Running xterm...")
		Spawn("xterm")
	}
	targetCode = C.XKeysymToKeycode(xm.backend.display, C.XStringToKeysym(C.CString("m")))
	if data.keycode == C.uint(targetCode) && cleanModifiers == C.uint(C.Mod4Mask) {
		fmt.Println("Win + m! Start moving windows!")
		currentMovingState := xm.startMoving.Load()
		xm.startMoving.Store(!currentMovingState)
	}
}

func (xm *XWindowManager) handleConfigureRequest(evWindow C.Window, data *C.XConfigureRequestEvent) {
	fmt.Println("XWM: Handling ConfigureRequest for window \"", evWindow, "\"")

	var changes C.XWindowChanges
	changes.x = data.x
	changes.y = data.y
	changes.width = data.width
	changes.height = data.height
	changes.border_width = data.border_width
	changes.sibling = data.above
	changes.stack_mode = data.detail

	clientContext, err := xm.clientManager.FindByClient(evWindow)

	if err != nil || clientContext == nil || clientContext.Frame == C.None {
		C.XConfigureWindow(xm.backend.display, evWindow, C.uint(data.value_mask), &changes)
		xm.backend.Flush()
		return
	}

	changes.width = data.width + TitlebarWidth

	C.XConfigureWindow(xm.backend.display, clientContext.Frame, C.uint(data.value_mask), &changes)

	C.XMoveResizeWindow(
		xm.backend.display,
		clientContext.Window,
		C.int(TitlebarWidth),
		0,
		C.uint(data.width),
		C.uint(data.height),
	)

	xm.backend.Flush()
}

func (xm *XWindowManager) handleExpose(evWindow C.Window, data *C.XExposeEvent) {
	fmt.Println("XWindowManager:Event:expose for window \"", evWindow, "\"")

	clientWin, err := xm.clientManager.FindByFrame(evWindow)
	if err != nil {
		fmt.Println("XWindowManager:Event:expose. Cannot find client window by frame:", err.Error())
		return
	}

	var rootReturn C.Window
	var x, y C.int
	var width, height, border C.uint
	var depth C.uint
	C.XGetGeometry(xm.backend.display, C.Drawable(evWindow), &rootReturn, &x, &y, &width, &height, &border, &depth)

	// create grafic context for drawing
	gc := C.XCreateGC(xm.backend.display, C.Drawable(evWindow), 0, nil)
	defer C.XFreeGC(xm.backend.display, gc)

	// drawing rectable
	// using color 0x3b82f6
	const backgroundColor uint64 = 0x3B82F6
	C.XSetForeground(xm.backend.display, gc, C.ulong(backgroundColor))
	C.XFillRectangle(xm.backend.display, C.Drawable(evWindow), gc, 0, 0, TitlebarWidth, height)

	// draw a text
	C.XSetForeground(xm.backend.display, gc, 0xFFFFFF)
	startY := C.int(20)
	charSpacing := C.int(12)
	textX := C.int(6)

	runes := []rune(clientWin.Title)

	for i, r := range runes {
		currentY := startY + C.int(i)*charSpacing

		if currentY > C.int(height)-10 {
			break
		}

		charStr := C.CString(string(r))
		C.XDrawString(xm.backend.display, C.Drawable(evWindow), gc, textX, currentY, charStr, 1)
		C.free(unsafe.Pointer(charStr))
	}

}

func (xm *XWindowManager) handleButtonPress(evWindow C.Window, data *C.XButtonEvent) {
	fmt.Println("XWM: Handling ButtonPress for window \"", evWindow, "\"")
	if clientContext, err := xm.clientManager.FindByFrame(evWindow); err == nil && clientContext != nil {
		if data.x > TitlebarWidth {
			fmt.Println("Client zone click, skip")
			return
		}

		clientContext.Motion.StartMouseX = data.x_root
		clientContext.Motion.StartMouseY = data.y_root

		C.XGetWindowAttributes(xm.backend.display, clientContext.Frame, &clientContext.Motion.StartAttr)

		xm.clientManager.SetMouseMovedClient(clientContext)

		C.XGrabPointer(
			xm.backend.display,
			data.window,
			C.False,
			C.uint(C.PointerMotionMask|C.ButtonReleaseMask),
			C.GrabModeAsync,
			C.GrabModeAsync,
			C.None,
			C.None,
			C.CurrentTime,
		)
	}
}
func (xm *XWindowManager) handleButtonRelease(evWindow C.Window, data *C.XButtonEvent) {
	fmt.Println("XWM: Handling ButtonRelease for window \"", evWindow, "\"")
	if clientContext := xm.clientManager.GetMouseMovedClient(); clientContext != nil {
		var child C.Window
		var lastX, lastY C.int
		C.XTranslateCoordinates(xm.backend.display, evWindow, xm.rootWindow, 0, 0, &lastX, &lastY, &child)
		C.XUngrabPointer(xm.backend.display, C.CurrentTime)
		// update geometry - setting new X,Y positions
		clientContext.Geom.X = int(lastX)
		clientContext.Geom.Y = int(lastY)
		xm.clientManager.SetMouseMovedClient(nil)
		xm.backend.Flush()
	}
}
func (xm *XWindowManager) handleMotionEvent(evWindow C.Window, data *C.XMotionEvent) {
	fmt.Println("XWM: Handling MotionEvent for window \"", evWindow, "\"")
	if clientContext := xm.clientManager.GetMouseMovedClient(); clientContext != nil {
		deltaX := data.x_root - clientContext.Motion.StartMouseX
		deltaY := data.y_root - clientContext.Motion.StartMouseY

		newX := clientContext.Motion.StartAttr.x + deltaX
		newY := clientContext.Motion.StartAttr.y + deltaY
		// update velocity accordance to mouse event
		clientContext.Motion.VelocityX = int(newX-C.int(clientContext.Geom.X)) % 7
		clientContext.Motion.VelocityY = int(newY-C.int(clientContext.Geom.Y)) % 7
		C.XMoveWindow(xm.backend.display, clientContext.Frame, newX, newY)
		xm.backend.Flush()
	}
}
