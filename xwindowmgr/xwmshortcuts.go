package xwindowmgr

/*
#include <X11/Xlib.h>
#include <X11/keysym.h>

*/
import "C"
import "fmt"

func (xm *XWindowManager) GrabShortcuts() {
	keyCode := xm.backend.KeysymToKeycode(C.KeySym(C.XK_Return))
	modifier := C.uint(C.Mod4Mask)
	xm.backend.GrabKey(xm.rootWindow, C.int(keyCode), modifier)
	keyCode = xm.backend.KeysymToKeycode(xm.backend.StringToKeysym("m"))
	if keyCode == 0 {
		fmt.Println("XWM: Error: Cannot grab a keyCode for \"m\"")
		return
	}
	xm.backend.GrabKey(xm.rootWindow, C.int(keyCode), modifier)
	xm.backend.Flush()
	fmt.Println("XWM:Hotkeys are registrated")
}
