package xwindowmgr

/*
#include <X11/Xlib.h>
#include "x11gohelper.hpp"
*/
import "C"
import (
	"math/rand/v2"
	"unsafe"
)

const TitlebarWidth = 20

type WindowRect struct {
	X      int
	Y      int
	Width  int
	Height int
}

type MotionData struct {
	IsMouseMoving bool
	StartMouseX   C.int
	StartMouseY   C.int
	StartAttr     C.XWindowAttributes

	VelocityX int
	VelocityY int
}

// Window controlled by WM
type Client struct {
	Title  string
	Window C.Window   // Descriptor of window
	Frame  C.Window   // Descriptor of window header - frame
	Geom   WindowRect // Current geometry

	Motion MotionData
}

func generateClientContext(backend *XBackend, window C.Window) *Client {
	var ret Client
	ret.Window = window

	geometry := backend.GetRect(ret.Window)
	if geometry != nil {
		ret.Geom = *geometry
		cName := C.get_window_name(backend.display, ret.Window)
		if cName != nil {
			ret.Title = C.GoString(cName)
			C.free(unsafe.Pointer(cName))
		} else {
			ret.Title = "Window"
		}
		ret.Motion.VelocityX = rand.IntN(10) - 5
		ret.Motion.VelocityY = rand.IntN(10) - 5
		return &ret
	}

	return nil
}

func (cl *Client) CreateFrame(backend *XBackend, rootWin C.Window) {

	frameWidth := cl.Geom.Width + TitlebarWidth
	frameWin := C.XCreateSimpleWindow(
		backend.display, rootWin,
		C.int(cl.Geom.X), C.int(cl.Geom.Y), C.uint(frameWidth), C.uint(cl.Geom.Height),
		1, C.XBlackPixel(backend.display, backend.screen), C.XWhitePixel(backend.display, backend.screen),
	)

	cl.Frame = frameWin
	// adjascent for exposure
	C.XSelectInput(backend.display, frameWin, C.ExposureMask|C.ButtonPressMask|C.SubstructureNotifyMask)

	// reparent the window
	C.XReparentWindow(backend.display, cl.Window, frameWin, TitlebarWidth, 0)
}
