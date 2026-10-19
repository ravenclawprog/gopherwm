package xwindowmgr

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/keysym.h>
#include "x11gohelper.hpp"

*/
import "C"
import "fmt"

func (xm *XWindowManager) GrabShortcuts() {

	keyCode := C.XKeysymToKeycode(xm.backend.display, C.KeySym(C.XK_Return))

	// mask for win key
	modifier := C.uint(C.Mod4Mask)

	// win + enter grub
	C.XGrabKey(
		xm.backend.display,
		C.int(keyCode),
		modifier,
		xm.rootWindow,
		C.True, // owner_events
		C.GrabModeAsync,
		C.GrabModeAsync,
	)

	xm.backend.Flush()
	fmt.Println("XWM:Hotkeys are registrated")
}
