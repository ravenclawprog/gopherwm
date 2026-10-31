package xwindowmgr

import "math/rand/v2"

const TitlebarWidth = 20

type WindowRect struct {
	X      int
	Y      int
	Width  int
	Height int
}

type MotionData struct {
	StartMouseX int
	StartMouseY int
	StartAttr   WindowAttributes

	VelocityX int
	VelocityY int
}

// Window controlled by WM
type Client struct {
	Title  string
	Window WindowID   // Descriptor of window
	Frame  WindowID   // Descriptor of window header - frame
	Geom   WindowRect // Current geometry

	Motion MotionData
}

func generateClientContext(backend Backend, window WindowID) *Client {
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

func (cl *Client) CreateFrame(backend Backend, rootWin WindowID) {

	frameWidth := cl.Geom.Width + TitlebarWidth
	cl.Frame = backend.CreateFrame(rootWin, cl.Geom.X, cl.Geom.Y, frameWidth, cl.Geom.Height)
	backend.Reparent(cl.Window, cl.Frame, TitlebarWidth, 0)
}
