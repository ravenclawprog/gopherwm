package xwindowmgr

/*
#include <X11/Xlib.h>
*/
import "C"
import "math/rand/v2"

const TitlebarWidth = 20

type WindowRect struct {
	X      int
	Y      int
	Width  int
	Height int
}

type MotionData struct {
	StartMouseX C.int
	StartMouseY C.int
	StartAttr   C.XWindowAttributes

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
		ret.Title = backend.GetWindowName(ret.Window)
		ret.Motion.VelocityX = rand.IntN(10) - 5
		ret.Motion.VelocityY = rand.IntN(10) - 5
		return &ret
	}

	return nil
}

func (cl *Client) CreateFrame(backend *XBackend, rootWin C.Window) {

	frameWidth := cl.Geom.Width + TitlebarWidth
	cl.Frame = backend.CreateFrame(rootWin, C.int(cl.Geom.X), C.int(cl.Geom.Y), C.int(frameWidth), C.int(cl.Geom.Height))
	backend.Reparent(cl.Window, cl.Frame, TitlebarWidth, 0)
}
