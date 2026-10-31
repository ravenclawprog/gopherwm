package xwindowmgr

import "fmt"

func (xm *XWindowManager) GrabShortcuts() {
	keyCode := xm.backend.KeysymToKeycode(KeyReturn)
	modifier := Mod4Mask
	xm.backend.GrabKey(xm.rootWindow, keyCode, modifier)
	keyCode = xm.backend.KeysymToKeycode(xm.backend.StringToKeysym("m"))
	if keyCode == 0 {
		fmt.Println("XWM: Error: Cannot grab a keyCode for \"m\"")
		return
	}
	xm.backend.GrabKey(xm.rootWindow, keyCode, modifier)
	xm.backend.Flush()
	fmt.Println("XWM:Hotkeys are registrated")
}
