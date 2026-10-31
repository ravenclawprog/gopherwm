package xwindowmgr

/*
#include <X11/Xlib.h>
#include <X11/keysym.h>
*/
import "C"
import "fmt"

func (xm *XWindowManager) handleMapRequest(evWindow C.Window, data *C.XMapRequestEvent) {
	fmt.Println("XWM: Handling MapRequest for window \"", evWindow, "\"")

	clientContext, _ := xm.clientManager.FindByClient(evWindow)

	if clientContext != nil && clientContext.Frame != C.None {
		xm.backend.MapWindow(clientContext.Frame)
		xm.backend.MapWindow(clientContext.Window)
	} else {
		newClient := generateClientContext(xm.backend, evWindow)
		newClient.CreateFrame(xm.backend, xm.rootWindow)
		fmt.Println("XWM: Handling MapRequest for window \"", evWindow, "\"; create frame window \"", newClient.Frame, "\"; Window Title:", newClient.Title)
		xm.clientManager.Add(newClient)
		xm.backend.MapWindow(newClient.Frame)
		xm.backend.MapWindow(newClient.Window)
	}
	xm.backend.Flush()
}

func (xm *XWindowManager) handleUnmapNotify(evWindow C.Window, data *C.XUnmapEvent) {
	fmt.Println("XWM: Handling UnmapNotify for window \"", evWindow, "\"")

	var targetFrame C.Window = C.None
	clientContext, err := xm.clientManager.FindByClient(evWindow)
	if err != nil {
		fmt.Println("XWM: Handling UnmapNotify: cannot find clientWindow:", err.Error())
		return
	}
	targetFrame = clientContext.Frame
	if targetFrame != C.None {
		xm.backend.UnmapSubwindows(clientContext.Frame)
		xm.backend.UnmapWindow(clientContext.Frame)
		xm.backend.Flush()
	}
}

func (xm *XWindowManager) handleDestroyNotify(evWindow C.Window, data *C.XDestroyWindowEvent) {
	fmt.Println("XWM: Handling Destroy Notify for window \"", evWindow, "\"")
	var targetFrame C.Window = C.None
	clientContext, err := xm.clientManager.FindByClient(evWindow)
	if err != nil {
		fmt.Println("XWM: Handling Destroy Notify: cannot find clientWindow:", err.Error())
		return
	}
	targetFrame = clientContext.Frame
	if targetFrame != C.None {
		xm.backend.UnmapSubwindows(targetFrame)
		xm.backend.UnmapWindow(targetFrame)
		xm.backend.DestroyWindow(targetFrame)
	}
	xm.clientManager.Remove(clientContext.Window)
	xm.backend.Flush()
}

func (xm *XWindowManager) handleKeyPress(evWindow C.Window, data *C.XKeyEvent) {
	fmt.Println("XWM: Handling KeyPress for window \"", evWindow, "\"")

	if data.subwindow != C.None {
		xm.backend.RaiseWindow(data.subwindow)
	}
	cleanModifiers := data.state &^ C.uint(C.Mod2Mask|C.LockMask)
	targetCode := xm.backend.KeysymToKeycode(C.KeySym(C.XK_Return))
	if data.keycode == C.uint(targetCode) && cleanModifiers == C.uint(C.Mod4Mask) {
		fmt.Println("Win + Enter! Running xterm...")
		Spawn("xterm")
	}
	targetCode = xm.backend.KeysymToKeycode(xm.backend.StringToKeysym("m"))
	if data.keycode == C.uint(targetCode) && cleanModifiers == C.uint(C.Mod4Mask) {
		fmt.Println("Win + m! Start moving windows!")
		currentMovingState := xm.startMoving.Load()
		xm.startMoving.Store(!currentMovingState)
	}
}

func (xm *XWindowManager) handleConfigureRequest(evWindow C.Window, data *C.XConfigureRequestEvent) {
	fmt.Println("XWM: Handling ConfigureRequest for window \"", evWindow, "\"")

	var changes C.XWindowChanges
	changes.x = data.x
	changes.y = data.y
	changes.width = data.width
	changes.height = data.height
	changes.border_width = data.border_width
	changes.sibling = data.above
	changes.stack_mode = data.detail

	clientContext, err := xm.clientManager.FindByClient(evWindow)

	if err != nil || clientContext == nil || clientContext.Frame == C.None {
		xm.backend.ConfigureWindow(evWindow, C.uint(data.value_mask), &changes)
		xm.backend.Flush()
		return
	}

	changes.width = data.width + TitlebarWidth

	xm.backend.ConfigureWindow(clientContext.Frame, C.uint(data.value_mask), &changes)

	xm.backend.MoveResizeWindow(
		clientContext.Window,
		C.int(TitlebarWidth),
		0,
		C.uint(data.width),
		C.uint(data.height),
	)

	xm.backend.Flush()
}

func (xm *XWindowManager) handleExpose(evWindow C.Window, data *C.XExposeEvent) {
	fmt.Println("XWindowManager:Event:expose for window \"", evWindow, "\"")

	clientWin, err := xm.clientManager.FindByFrame(evWindow)
	if err != nil {
		fmt.Println("XWindowManager:Event:expose. Cannot find client window by frame:", err.Error())
		return
	}

	rect := xm.backend.GetRect(evWindow)
	if rect == nil {
		return
	}
	height := C.uint(rect.Height)
	gc := xm.backend.CreateGC(evWindow)
	defer xm.backend.FreeGC(gc)

	// drawing rectable
	// using color 0x3b82f6
	const backgroundColor uint64 = 0x3B82F6
	xm.backend.SetForeground(gc, backgroundColor)
	xm.backend.FillRectangle(evWindow, gc, 0, 0, TitlebarWidth, height)

	// draw a text
	xm.backend.SetForeground(gc, 0xFFFFFF)
	startY := C.int(20)
	charSpacing := C.int(12)
	textX := C.int(6)

	runes := []rune(clientWin.Title)

	for i, r := range runes {
		currentY := startY + C.int(i)*charSpacing

		if currentY > C.int(height)-10 {
			break
		}

		xm.backend.DrawString(evWindow, gc, textX, currentY, string(r), 1)
	}

}

func (xm *XWindowManager) handleButtonPress(evWindow C.Window, data *C.XButtonEvent) {
	fmt.Println("XWM: Handling ButtonPress for window \"", evWindow, "\"")
	if clientContext, err := xm.clientManager.FindByFrame(evWindow); err == nil && clientContext != nil {
		if data.x > TitlebarWidth {
			fmt.Println("Client zone click, skip")
			return
		}

		clientContext.Motion.StartMouseX = data.x_root
		clientContext.Motion.StartMouseY = data.y_root

		clientContext.Motion.StartAttr = xm.backend.GetWindowAttributes(clientContext.Frame)

		xm.clientManager.SetMouseMovedClient(clientContext)

		xm.backend.GrabPointer(data.window)
	}
}
func (xm *XWindowManager) handleButtonRelease(evWindow C.Window, data *C.XButtonEvent) {
	fmt.Println("XWM: Handling ButtonRelease for window \"", evWindow, "\"")
	if clientContext := xm.clientManager.GetMouseMovedClient(); clientContext != nil {
		lastX, lastY := xm.backend.TranslateCoordinates(evWindow, xm.rootWindow, 0, 0)
		xm.backend.UngrabPointer()
		// update geometry - setting new X,Y positions
		clientContext.Geom.X = int(lastX)
		clientContext.Geom.Y = int(lastY)
		xm.clientManager.SetMouseMovedClient(nil)
		xm.backend.Flush()
	}
}
func (xm *XWindowManager) handleMotionEvent(evWindow C.Window, data *C.XMotionEvent) {
	fmt.Println("XWM: Handling MotionEvent for window \"", evWindow, "\"")
	if clientContext := xm.clientManager.GetMouseMovedClient(); clientContext != nil {
		deltaX := data.x_root - clientContext.Motion.StartMouseX
		deltaY := data.y_root - clientContext.Motion.StartMouseY

		newX := clientContext.Motion.StartAttr.x + deltaX
		newY := clientContext.Motion.StartAttr.y + deltaY
		// update velocity accordance to mouse event
		clientContext.Motion.VelocityX = int(newX-C.int(clientContext.Geom.X)) % 7
		clientContext.Motion.VelocityY = int(newY-C.int(clientContext.Geom.Y)) % 7
		xm.backend.MoveWindow(clientContext.Frame, newX, newY)
		xm.backend.Flush()
	}
}
