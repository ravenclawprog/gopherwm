package main

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include "x11gohelper.hpp"

#define MAX(a, b) ((a) > (b) ? (a) : (b))
*/
import "C"
import (
	"fmt"
	"gopherwm/idgen"
	"math/rand/v2"
	"slices"
	"time"
	"unsafe"
)

var windowSlice []WindowMove

type WindowRect struct {
	x, y          C.int
	width, height C.uint
}

type WindowMove struct {
	window       C.Window
	rect         WindowRect
	vel_x, vel_y C.int
}

func unionToInt(cbytes [192]byte) (result int) {
	var addr *byte = &cbytes[0]
	var cast *C.int = (*C.int)(unsafe.Pointer(addr))
	return (int)(*cast)
}

func unionToXKeyEvent(cbytes [192]byte) (result *C.XKeyEvent) {
	var addr *byte = &cbytes[0]
	var cast *C.XKeyEvent = (*C.XKeyEvent)(unsafe.Pointer(addr))
	return cast
}

func unionToXMapEvent(cbytes [192]byte) (result *C.XMapEvent) {
	var addr *byte = &cbytes[0]
	var cast *C.XMapEvent = (*C.XMapEvent)(unsafe.Pointer(addr))
	return cast
}

func unionToXButtonEvent(cbytes [192]byte) (result *C.XButtonEvent) {
	var addr *byte = &cbytes[0]
	var cast *C.XButtonEvent = (*C.XButtonEvent)(unsafe.Pointer(addr))
	return cast
}

func getWindowGeometry(dpy *C.Display, root C.Window, rootRect *WindowRect) {
	var out_root C.Window
	var out_border_width C.uint
	var out_depth C.uint

	C.XGetGeometry(dpy, root, &out_root, &(rootRect.x), &(rootRect.y), &(rootRect.width), &(rootRect.height), &out_border_width, &out_depth)
}

func addWindow(dpy *C.Display, wind *C.Window, s *[]WindowMove) {
	fmt.Printf("Adding window\n")
	if s == nil {
		return
	}
	var move WindowMove
	move.window = *wind
	getWindowGeometry(dpy, *wind, &move.rect)
	move.vel_x = C.int(rand.IntN(10) - 5)
	move.vel_y = C.int(rand.IntN(10) - 5)
	*s = append(*s, move)
}

func removeWindow(wind *C.Window, s *[]WindowMove) {
	fmt.Printf("Removing window\n")
	if s == nil {
		return
	}

	index := -1
	for i, v := range *s {
		if v.window == *wind {
			index = i
			break
		}
	}

	slices.Delete(*s, index, index+1)
}

func addAllWindows(display *C.Display, root *C.Window) {
	var (
		root_return      C.Window
		parent_return    C.Window
		children_return  *C.Window
		nchildren_return C.uint
	)

	C.XQueryTree(display, *root, &root_return, &parent_return, &children_return, &nchildren_return)
	currentRawPtr := unsafe.Pointer(children_return)
	fmt.Printf("number of windows is %d\n", nchildren_return)
	for i := 0; i < int(nchildren_return); i++ {
		addWindow(display, (*C.Window)(currentRawPtr), &windowSlice)
		currentRawPtr = unsafe.Pointer(uintptr(currentRawPtr) + unsafe.Sizeof(children_return))
	}

	C.XFree(unsafe.Pointer(children_return))
}
func moveWindow(display *C.Display, move *WindowMove, rootRect *WindowRect) {
	fmt.Printf("Move window:")
	var rect *WindowRect = &move.rect
	rect.x = rect.x + move.vel_x
	rect.y = rect.y + move.vel_y
	if C.int(rect.x)+C.int(rect.width) > C.int(rootRect.width) || (rect.x < 0) {
		move.vel_x *= -1
	}
	if ((C.int(rect.y) + C.int(rect.height)) > C.int(rootRect.height)) || (rect.y < 0) {
		move.vel_y *= -1
	}

	C.XMoveWindow(display, move.window, rect.x, rect.y)
}

func moveAllWindows(display *C.Display, rootRect *WindowRect, windowMoveSlice *[]WindowMove) {
	for i := range *windowMoveSlice {
		moveWindow(display, &(*windowMoveSlice)[i], rootRect)
	}
}

func main() {
	var dpy *C.Display
	var attr C.XWindowAttributes
	var start C.XButtonEvent
	var ev C.XEvent
	var rootRect WindowRect

	id_gen := idgen.NewIDGenerator()
	fmt.Println("generated data:", id_gen.Generate())

	fmt.Printf("WM start\n")
	dpy = C.XOpenDisplay(nil)
	if dpy == nil {
		fmt.Printf("Cannot open display, out\n")
		return
	}
	fmt.Printf("Getting root...\n")
	root := C._DefaultRootWindow(dpy)
	getWindowGeometry(dpy, root, &rootRect)
	addAllWindows(dpy, &root)
	fmt.Printf("Grabbing keys for moving and resizing...\n")
	C.XGrabKey(
		dpy,
		C.int(C.XKeysymToKeycode(dpy, C.XStringToKeysym(C.CString("F1")))),
		C.Mod1Mask,
		root,
		1,
		C.GrabModeAsync,
		C.GrabModeAsync,
	)

	C.XGrabButton(
		dpy,
		C.Button1,
		C.Mod1Mask,
		root,
		1,
		C.ButtonPressMask|C.ButtonReleaseMask|C.PointerMotionMask,
		C.GrabModeAsync,
		C.GrabModeAsync,
		C.None,
		C.None,
	)

	C.XGrabButton(
		dpy,
		C.Button3,
		C.Mod1Mask,
		root,
		1,
		C.ButtonPressMask|C.ButtonReleaseMask|C.PointerMotionMask,
		C.GrabModeAsync,
		C.GrabModeAsync,
		C.None,
		C.None,
	)

	start.subwindow = C.None
	fmt.Printf("Select input...\n")
	var inputMask C.long = C.KeyPressMask | C.SubstructureNotifyMask
	C.XSelectInput(dpy, root, inputMask)
	for {
		for C.XPending(dpy) != 0 {
			C.XNextEvent(dpy, &ev)
			evType := unionToInt(ev)
			fmt.Printf("Event is:%d\n", evType)
			if evType == C.KeyPress {
				fmt.Printf("Key pressed:%d\n", unionToXKeyEvent(ev).keycode)
				if unionToXKeyEvent(ev).subwindow != C.None {
					C.XRaiseWindow(dpy, unionToXKeyEvent(ev).subwindow)
				}
			} else if evType == C.MapNotify {
				fmt.Printf("Event: MapNotify\n")
				addWindow(dpy, &(unionToXMapEvent(ev).window), &windowSlice)
			} else if evType == C.UnmapNotify {
				fmt.Printf("Event: UnmapNotify\n")
				removeWindow(&(unionToXMapEvent(ev).window), &windowSlice)
			} else if evType == C.ButtonPress {
				fmt.Printf("Button pressed\n")
				if unionToXButtonEvent(ev).subwindow != C.None {
					C.XGetWindowAttributes(dpy, unionToXButtonEvent(ev).subwindow, &attr)
					start = *unionToXButtonEvent(ev)
				}
			} else if evType == C.MotionNotify && start.subwindow != C.None {
				fmt.Printf("Motion notify\n")
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
					dpy,
					start.subwindow,
					attr.x+toDiffX,
					attr.y+toDiffY,
					max(1, attr.width+toWidth),
					max(1, attr.height+toHeight))
			} else if evType == C.ButtonRelease {
				fmt.Printf("Button release \n")
				start.subwindow = C.None
			}
		}
		moveAllWindows(dpy, &rootRect, &windowSlice)
		time.Sleep(30 * time.Millisecond)
	}
}

func max(a, b C.int) C.uint {
	if a > b {
		return C.uint(a)
	}

	return C.uint(b)
}
