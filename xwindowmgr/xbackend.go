package xwindowmgr

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include "x11gohelper.hpp"
*/
import "C"
import (
	"errors"
	"unsafe"
)

// These aliases expose backend signatures without requiring callers to import C.
// They retain Xlib's representation; replacing it with Go-owned types is separate work.
type WindowID = C.Window
type Coordinate = C.int
type Dimension = C.uint
type WindowAttributes = C.XWindowAttributes
type WindowChanges = C.XWindowChanges
type KeySym = C.KeySym
type KeyCode = C.KeyCode
type GraphicsContext = C.GC

type MapRequestEvent = C.XMapRequestEvent
type UnmapEvent = C.XUnmapEvent
type DestroyWindowEvent = C.XDestroyWindowEvent

const (
	MapRequest    = C.MapRequest
	UnmapNotify   = C.UnmapNotify
	DestroyNotify = C.DestroyNotify
)

type XBackend struct {
	display *C.Display
	screen  C.int
}

func NewXBackend() (*XBackend, error) {
	C.XInitThreads()

	display := C.XOpenDisplay(nil)
	if display == nil {
		return nil, errors.New("cannot open display")
	}
	C.SetupErrorHandler()

	return &XBackend{
		display: display,
		screen:  C.XDefaultScreen(display),
	}, nil
}

func (b *XBackend) GetRootWindow() C.Window {
	return C._DefaultRootWindow(b.display)
}

func (b *XBackend) GetRect(window C.Window) *WindowRect {
	var (
		outWindow      C.Window
		outBorderWidth C.uint
		outDepth       C.uint

		cX      C.int
		cY      C.int
		cWidth  C.uint
		cHeight C.uint
	)
	var ret *WindowRect = nil
	status := C.XGetGeometry(
		b.display,
		window,
		&outWindow,
		&cX,
		&cY,
		&cWidth,
		&cHeight,
		&outBorderWidth,
		&outDepth,
	)
	if status != 0 {
		ret = &WindowRect{}
		ret.X = int(cX)
		ret.Y = int(cY)
		ret.Width = int(cWidth)
		ret.Height = int(cHeight)
	}
	return ret
}

func (b *XBackend) SelectInput(inputMask int64) {
	C.XSelectInput(b.display, b.GetRootWindow(), C.long(inputMask))
}
func (b *XBackend) CreateFrame(parent C.Window, x, y, w, h C.int) C.Window {
	frameWin := C.XCreateSimpleWindow(
		b.display, parent,
		x, y, C.uint(w), C.uint(h),
		1, C.XBlackPixel(b.display, b.screen), C.XWhitePixel(b.display, b.screen),
	)
	C.XSelectInput(b.display, frameWin, C.ExposureMask|C.ButtonPressMask|C.SubstructureNotifyMask)
	return frameWin
}

func (b *XBackend) Reparent(client C.Window, frame C.Window, offsetX, offsetY C.int) {
	C.XReparentWindow(b.display, client, frame, offsetX, offsetY)
}

func (b *XBackend) DestroyWindow(win C.Window) {
	C.XDestroyWindow(b.display, win)
}

func (b *XBackend) Flush() {
	C.XFlush(b.display)
}

func (b *XBackend) Sync(discard bool) {
	if discard {
		C.XSync(b.display, C.True)
	} else {
		C.XSync(b.display, C.False)
	}
}

// GetWindowName returns a Go-owned copy and releases the Xlib allocation.
func (b *XBackend) GetWindowName(window C.Window) string {
	name := C.get_window_name(b.display, window)
	if name == nil {
		return "Window"
	}
	defer C.XFree(unsafe.Pointer(name))
	return C.GoString(name)
}

// QueryTree copies child handles before releasing the Xlib allocation.
func (b *XBackend) QueryTree(window C.Window) []C.Window {
	var root, parent C.Window
	var children *C.Window
	var count C.uint
	status := C.XQueryTree(b.display, window, &root, &parent, &children, &count)
	if children != nil {
		defer C.XFree(unsafe.Pointer(children))
	}
	if status == 0 || count == 0 {
		return nil
	}
	return append([]C.Window(nil), unsafe.Slice(children, int(count))...)
}

func (b *XBackend) GetWindowAttributes(window C.Window) C.XWindowAttributes {
	var attrs C.XWindowAttributes
	C.XGetWindowAttributes(b.display, window, &attrs)
	return attrs
}

func (b *XBackend) MapWindow(window C.Window)              { C.XMapWindow(b.display, window) }
func (b *XBackend) UnmapWindow(window C.Window)            { C.XUnmapWindow(b.display, window) }
func (b *XBackend) UnmapSubwindows(window C.Window)        { C.XUnmapSubwindows(b.display, window) }
func (b *XBackend) RaiseWindow(window C.Window)            { C.XRaiseWindow(b.display, window) }
func (b *XBackend) MoveWindow(window C.Window, x, y C.int) { C.XMoveWindow(b.display, window, x, y) }
func (b *XBackend) MoveResizeWindow(window C.Window, x, y C.int, width, height C.uint) {
	C.XMoveResizeWindow(b.display, window, x, y, width, height)
}
func (b *XBackend) ConfigureWindow(window C.Window, mask C.uint, changes *C.XWindowChanges) {
	C.XConfigureWindow(b.display, window, mask, changes)
}

func (b *XBackend) Pending() int { return int(C.XPending(b.display)) }
func (b *XBackend) NextEvent() WMEvent {
	var event C.XEvent
	C.XNextEvent(b.display, &event)
	return parseXEvent(event)
}

func (b *XBackend) KeysymToKeycode(keysym C.KeySym) C.KeyCode {
	return C.XKeysymToKeycode(b.display, keysym)
}
func (b *XBackend) StringToKeysym(name string) C.KeySym {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return C.XStringToKeysym(cName)
}
func (b *XBackend) GrabKey(window C.Window, keycode C.int, modifiers C.uint) {
	C.XGrabKey(b.display, keycode, modifiers, window, C.True, C.GrabModeAsync, C.GrabModeAsync)
}

// GrabPointer captures motion and release events during window dragging.
func (b *XBackend) GrabPointer(window C.Window) {
	C.XGrabPointer(b.display, window, C.False, C.uint(C.PointerMotionMask|C.ButtonReleaseMask),
		C.GrabModeAsync, C.GrabModeAsync, C.None, C.None, C.CurrentTime)
}
func (b *XBackend) UngrabPointer() { C.XUngrabPointer(b.display, C.CurrentTime) }
func (b *XBackend) TranslateCoordinates(source, destination C.Window, x, y C.int) (C.int, C.int) {
	var child C.Window
	var outX, outY C.int
	C.XTranslateCoordinates(b.display, source, destination, x, y, &outX, &outY, &child)
	return outX, outY
}

func (b *XBackend) CreateGC(window C.Window) C.GC {
	return C.XCreateGC(b.display, C.Drawable(window), 0, nil)
}
func (b *XBackend) FreeGC(gc C.GC) { C.XFreeGC(b.display, gc) }
func (b *XBackend) SetForeground(gc C.GC, color uint64) {
	C.XSetForeground(b.display, gc, C.ulong(color))
}
func (b *XBackend) FillRectangle(window C.Window, gc C.GC, x, y C.int, width, height C.uint) {
	C.XFillRectangle(b.display, C.Drawable(window), gc, x, y, width, height)
}

// DrawString preserves XDrawString's byte-oriented text semantics.
func (b *XBackend) DrawString(window C.Window, gc C.GC, x, y C.int, text string, length int) {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	C.XDrawString(b.display, C.Drawable(window), gc, x, y, cText, C.int(length))
}
