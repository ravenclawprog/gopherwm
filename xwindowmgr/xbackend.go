package xwindowmgr

/*
#include <X11/Xlib.h>
#include "x11gohelper.hpp"
*/
import "C"
import "errors"

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
