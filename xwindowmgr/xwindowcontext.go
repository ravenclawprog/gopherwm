package xwindowmgr

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
*/
import "C"

type xWindowRect struct {
	x, y          C.int
	width, height C.uint
}

type xWindowMovingParams struct {
	velocityX, velocityY int
}

type xWindowContext struct {
	windowHandle C.Window            // handle of XWindow system
	rect         xWindowRect         // current position and size of XWindow
	movingData   xWindowMovingParams // parameters for update
}
