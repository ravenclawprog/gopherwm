package xwindowmgr

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/keysym.h>
#include "x11gohelper.hpp"
*/
import "C"
import (
	"errors"
	"sync"
	"unsafe"
)

// Constants retain Go types while taking their values from Xlib headers.
const (
	NoWindow WindowID = C.None
)

const (
	KeyPress         EventType = C.KeyPress
	ButtonPress      EventType = C.ButtonPress
	ButtonRelease    EventType = C.ButtonRelease
	MotionNotify     EventType = C.MotionNotify
	Expose           EventType = C.Expose
	DestroyNotify    EventType = C.DestroyNotify
	UnmapNotify      EventType = C.UnmapNotify
	MapRequest       EventType = C.MapRequest
	ConfigureRequest EventType = C.ConfigureRequest
)

const (
	KeyPressMask             EventMask = C.KeyPressMask
	SubstructureRedirectMask EventMask = C.SubstructureRedirectMask
)

const (
	LockMask ModifierMask = C.LockMask
	Mod2Mask ModifierMask = C.Mod2Mask
	Mod4Mask ModifierMask = C.Mod4Mask
)

const (
	KeyReturn KeySym = C.XK_Return
)

const (
	CWX           ConfigureMask = C.CWX
	CWY           ConfigureMask = C.CWY
	CWWidth       ConfigureMask = C.CWWidth
	CWHeight      ConfigureMask = C.CWHeight
	CWBorderWidth ConfigureMask = C.CWBorderWidth
	CWSibling     ConfigureMask = C.CWSibling
	CWStackMode   ConfigureMask = C.CWStackMode
)

const (
	IsUnmapped   MapState = C.IsUnmapped
	IsUnviewable MapState = C.IsUnviewable
	IsViewable   MapState = C.IsViewable
)

type XBackend struct {
	display *C.Display
	screen  C.int

	// Native GC pointers never leave the backend. Public handles are Go IDs.
	gcMu        sync.Mutex
	contexts    map[GraphicsContext]C.GC
	nextContext GraphicsContext
}

func NewXBackend() (*XBackend, error) {
	C.XInitThreads()
	display := C.XOpenDisplay(nil)
	if display == nil {
		return nil, errors.New("cannot open display")
	}
	C.SetupErrorHandler()
	return &XBackend{
		display: display, screen: C.XDefaultScreen(display),
		contexts: make(map[GraphicsContext]C.GC),
	}, nil
}

func (b *XBackend) GetRootWindow() WindowID {
	return WindowID(C._DefaultRootWindow(b.display))
}

func (b *XBackend) GetRect(window WindowID) *WindowRect {
	var root C.Window
	var border, depth, width, height C.uint
	var x, y C.int
	if C.XGetGeometry(b.display, C.Drawable(window), &root, &x, &y, &width, &height, &border, &depth) == 0 {
		return nil
	}
	return &WindowRect{X: int(x), Y: int(y), Width: int(width), Height: int(height)}
}

func (b *XBackend) SelectInput(mask EventMask) {
	C.XSelectInput(b.display, C.Window(b.GetRootWindow()), C.long(mask))
}
func (b *XBackend) CreateFrame(parent WindowID, x, y, width, height int) WindowID {
	frame := C.XCreateSimpleWindow(b.display, C.Window(parent), C.int(x), C.int(y), C.uint(width), C.uint(height),
		1, C.XBlackPixel(b.display, b.screen), C.XWhitePixel(b.display, b.screen))
	C.XSelectInput(b.display, frame, C.ExposureMask|C.ButtonPressMask|C.SubstructureNotifyMask)
	return WindowID(frame)
}
func (b *XBackend) Reparent(client, frame WindowID, x, y int) {
	C.XReparentWindow(b.display, C.Window(client), C.Window(frame), C.int(x), C.int(y))
}
func (b *XBackend) DestroyWindow(window WindowID) { C.XDestroyWindow(b.display, C.Window(window)) }
func (b *XBackend) Flush()                        { C.XFlush(b.display) }
func (b *XBackend) Sync(discard bool) {
	flag := C.int(C.False)
	if discard {
		flag = C.True
	}
	C.XSync(b.display, flag)
}

// GetWindowName returns a Go-owned copy and releases the Xlib allocation.
func (b *XBackend) GetWindowName(window WindowID) string {
	name := C.get_window_name(b.display, C.Window(window))
	if name == nil {
		return "Window"
	}
	defer C.XFree(unsafe.Pointer(name))
	return C.GoString(name)
}

// QueryTree converts child IDs before releasing the Xlib allocation.
func (b *XBackend) QueryTree(window WindowID) []WindowID {
	var root, parent C.Window
	var children *C.Window
	var count C.uint
	status := C.XQueryTree(b.display, C.Window(window), &root, &parent, &children, &count)
	if children != nil {
		defer C.XFree(unsafe.Pointer(children))
	}
	if status == 0 || count == 0 {
		return nil
	}
	result := make([]WindowID, int(count))
	for i, child := range unsafe.Slice(children, int(count)) {
		result[i] = WindowID(child)
	}
	return result
}

func (b *XBackend) GetWindowAttributes(window WindowID) WindowAttributes {
	var attrs C.XWindowAttributes
	C.XGetWindowAttributes(b.display, C.Window(window), &attrs)
	return WindowAttributes{
		X: int(attrs.x), Y: int(attrs.y), Width: int(attrs.width), Height: int(attrs.height),
		MapState: MapState(attrs.map_state), OverrideRedirect: attrs.override_redirect != 0,
	}
}
func (b *XBackend) MapWindow(window WindowID)       { C.XMapWindow(b.display, C.Window(window)) }
func (b *XBackend) UnmapWindow(window WindowID)     { C.XUnmapWindow(b.display, C.Window(window)) }
func (b *XBackend) UnmapSubwindows(window WindowID) { C.XUnmapSubwindows(b.display, C.Window(window)) }
func (b *XBackend) RaiseWindow(window WindowID)     { C.XRaiseWindow(b.display, C.Window(window)) }
func (b *XBackend) MoveWindow(window WindowID, x, y int) {
	C.XMoveWindow(b.display, C.Window(window), C.int(x), C.int(y))
}
func (b *XBackend) MoveResizeWindow(window WindowID, x, y int, width, height uint) {
	C.XMoveResizeWindow(b.display, C.Window(window), C.int(x), C.int(y), C.uint(width), C.uint(height))
}
func (b *XBackend) ConfigureWindow(window WindowID, mask ConfigureMask, changes *WindowChanges) {
	native := C.XWindowChanges{
		x: C.int(changes.X), y: C.int(changes.Y), width: C.int(changes.Width), height: C.int(changes.Height),
		border_width: C.int(changes.BorderWidth), sibling: C.Window(changes.Sibling), stack_mode: C.int(changes.StackMode),
	}
	C.XConfigureWindow(b.display, C.Window(window), C.uint(mask), &native)
}
func (b *XBackend) Pending() int { return int(C.XPending(b.display)) }
func (b *XBackend) NextEvent() WMEvent {
	var event C.XEvent
	C.XNextEvent(b.display, &event)
	return parseXEvent(event)
}

// parseXEvent copies only manager-facing fields out of the native union.
// No native pointers or references to the event buffer escape into WMEvent.
func parseXEvent(event C.XEvent) WMEvent {
	ptr := unsafe.Pointer(&event)
	result := WMEvent{Type: EventType((*C.XAnyEvent)(ptr)._type)}
	switch result.Type {
	case MapRequest:
		data := (*C.XMapRequestEvent)(ptr)
		result.Window = WindowID(data.window)
		result.Data = MapRequestEvent{Parent: WindowID(data.parent)}
	case UnmapNotify:
		data := (*C.XUnmapEvent)(ptr)
		result.Window = WindowID(data.window)
		result.Data = UnmapEvent{FromConfigure: data.from_configure != 0}
	case DestroyNotify:
		data := (*C.XDestroyWindowEvent)(ptr)
		result.Window = WindowID(data.window)
		result.Data = DestroyWindowEvent{}
	case ConfigureRequest:
		data := (*C.XConfigureRequestEvent)(ptr)
		result.Window = WindowID(data.window)
		result.Data = ConfigureRequestEvent{
			X: int(data.x), Y: int(data.y), Width: int(data.width), Height: int(data.height),
			BorderWidth: int(data.border_width), Above: WindowID(data.above), Detail: int(data.detail),
			ValueMask: ConfigureMask(data.value_mask),
		}
	case KeyPress:
		data := (*C.XKeyEvent)(ptr)
		result.Window = WindowID(data.window)
		result.Data = KeyEvent{Subwindow: WindowID(data.subwindow), State: ModifierMask(data.state), Keycode: KeyCode(data.keycode)}
	case Expose:
		data := (*C.XExposeEvent)(ptr)
		result.Window = WindowID(data.window)
		result.Data = ExposeEvent{X: int(data.x), Y: int(data.y), Width: int(data.width), Height: int(data.height), Count: int(data.count)}
	case ButtonPress, ButtonRelease:
		data := (*C.XButtonEvent)(ptr)
		result.Window = WindowID(data.window)
		result.Data = ButtonEvent{
			X: int(data.x), Y: int(data.y), XRoot: int(data.x_root), YRoot: int(data.y_root),
			State: ModifierMask(data.state), Button: uint(data.button),
		}
	case MotionNotify:
		data := (*C.XMotionEvent)(ptr)
		result.Window = WindowID(data.window)
		result.Data = MotionEvent{X: int(data.x), Y: int(data.y), XRoot: int(data.x_root), YRoot: int(data.y_root), State: ModifierMask(data.state)}
	}
	return result
}

func (b *XBackend) KeysymToKeycode(keysym KeySym) KeyCode {
	return KeyCode(C.XKeysymToKeycode(b.display, C.KeySym(keysym)))
}
func (b *XBackend) StringToKeysym(name string) KeySym {
	native := C.CString(name)
	defer C.free(unsafe.Pointer(native))
	return KeySym(C.XStringToKeysym(native))
}
func (b *XBackend) GrabKey(window WindowID, keycode KeyCode, modifiers ModifierMask) {
	C.XGrabKey(b.display, C.int(keycode), C.uint(modifiers), C.Window(window), C.True, C.GrabModeAsync, C.GrabModeAsync)
}
func (b *XBackend) GrabPointer(window WindowID) {
	C.XGrabPointer(b.display, C.Window(window), C.False, C.uint(C.PointerMotionMask|C.ButtonReleaseMask),
		C.GrabModeAsync, C.GrabModeAsync, C.None, C.None, C.CurrentTime)
}
func (b *XBackend) UngrabPointer() { C.XUngrabPointer(b.display, C.CurrentTime) }
func (b *XBackend) TranslateCoordinates(source, destination WindowID, x, y int) (int, int) {
	var child C.Window
	var outX, outY C.int
	C.XTranslateCoordinates(b.display, C.Window(source), C.Window(destination), C.int(x), C.int(y), &outX, &outY, &child)
	return int(outX), int(outY)
}

func (b *XBackend) CreateGC(window WindowID) GraphicsContext {
	native := C.XCreateGC(b.display, C.Drawable(window), 0, nil)
	if native == nil {
		return 0
	}
	b.gcMu.Lock()
	defer b.gcMu.Unlock()
	b.nextContext++
	handle := b.nextContext
	b.contexts[handle] = native
	return handle
}
func (b *XBackend) FreeGC(handle GraphicsContext) {
	b.gcMu.Lock()
	defer b.gcMu.Unlock()
	if native := b.contexts[handle]; native != nil {
		C.XFreeGC(b.display, native)
		delete(b.contexts, handle)
	}
}
func (b *XBackend) SetForeground(handle GraphicsContext, color uint64) {
	b.gcMu.Lock()
	defer b.gcMu.Unlock()
	if native := b.contexts[handle]; native != nil {
		C.XSetForeground(b.display, native, C.ulong(color))
	}
}
func (b *XBackend) FillRectangle(window WindowID, handle GraphicsContext, x, y int, width, height uint) {
	b.gcMu.Lock()
	defer b.gcMu.Unlock()
	if native := b.contexts[handle]; native != nil {
		C.XFillRectangle(b.display, C.Drawable(window), native, C.int(x), C.int(y), C.uint(width), C.uint(height))
	}
}

// DrawString preserves XDrawString's byte-oriented text semantics.
func (b *XBackend) DrawString(window WindowID, handle GraphicsContext, x, y int, text string, length int) {
	b.gcMu.Lock()
	defer b.gcMu.Unlock()
	if native := b.contexts[handle]; native != nil {
		str := C.CString(text)
		defer C.free(unsafe.Pointer(str))
		C.XDrawString(b.display, C.Drawable(window), native, C.int(x), C.int(y), str, C.int(length))
	}
}
