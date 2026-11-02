package xwindowmgr

import "fmt"

func (xm *XWindowManager) GrabShortcuts() {
	keyCode := xm.backend.KeysymToKeycode(KeyReturn)
	if keyCode == 0 {
		fmt.Println("XWM: Error: Cannot grab a keyCode for \"m\"")
		return
	}
	xm.grabShortcut(keyCode)
	keyCode = xm.backend.KeysymToKeycode(xm.backend.StringToKeysym("m"))
	if keyCode == 0 {
		fmt.Println("XWM: Error: Cannot grab a keyCode for \"m\"")
		return
	}
	xm.grabShortcut(keyCode)
	xm.backend.Flush()
	fmt.Println("XWM:Hotkeys are registrated")
}

func (xm *XWindowManager) grabShortcut(keyCode KeyCode) {
	// Passive grabs require an exact modifier match, including lock states.
	// Keep these variants consistent with the masks ignored in handleKeyPress.
	for _, locks := range []ModifierMask{0, LockMask, Mod2Mask, LockMask | Mod2Mask} {
		xm.backend.GrabKey(xm.rootWindow, keyCode, Mod4Mask|locks)
	}
}
