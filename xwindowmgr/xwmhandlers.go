package xwindowmgr

import (
	"fmt"
	"math"
)

func (xm *XWindowManager) handleMapRequest(evWindow WindowID, data *MapRequestEvent) {
	fmt.Println("XWM: Handling MapRequest for window \"", evWindow, "\"")

	clientContext, _ := xm.clientManager.FindByClient(evWindow)

	if clientContext != nil && clientContext.Frame != NoWindow {
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

func (xm *XWindowManager) handleUnmapNotify(evWindow WindowID, data *UnmapEvent) {
	fmt.Println("XWM: Handling UnmapNotify for window \"", evWindow, "\"")

	var targetFrame WindowID = NoWindow
	clientContext, err := xm.clientManager.FindByClient(evWindow)
	if err != nil {
		fmt.Println("XWM: Handling UnmapNotify: cannot find clientWindow:", err.Error())
		return
	}
	targetFrame = clientContext.Frame
	if targetFrame != NoWindow {
		xm.backend.UnmapSubwindows(clientContext.Frame)
		xm.backend.UnmapWindow(clientContext.Frame)
		xm.backend.Flush()
	}
}

func (xm *XWindowManager) handleDestroyNotify(evWindow WindowID, data *DestroyWindowEvent) {
	fmt.Println("XWM: Handling Destroy Notify for window \"", evWindow, "\"")
	var targetFrame WindowID = NoWindow
	clientContext, err := xm.clientManager.FindByClient(evWindow)
	if err != nil {
		fmt.Println("XWM: Handling Destroy Notify: cannot find clientWindow:", err.Error())
		return
	}
	targetFrame = clientContext.Frame
	if targetFrame != NoWindow {
		xm.backend.UnmapSubwindows(targetFrame)
		xm.backend.UnmapWindow(targetFrame)
		xm.backend.DestroyWindow(targetFrame)
	}
	xm.clientManager.Remove(clientContext.Window)
	xm.backend.Flush()
}

func (xm *XWindowManager) handleKeyPress(evWindow WindowID, data *KeyEvent) {
	fmt.Println("XWM: Handling KeyPress for window \"", evWindow, "\"")

	if data.Subwindow != NoWindow {
		xm.backend.RaiseWindow(data.Subwindow)
	}

	cleanModifiers := data.State &^ (Mod2Mask | LockMask)
	key := shortcutKey{
		keycode:   data.Keycode,
		modifiers: cleanModifiers,
	}

	action, exists := xm.shortcuts[key]
	if !exists {
		return
	}

	switch action.action {
	case "spawn":
		Spawn(action.command[0], action.command[1:]...)
	case "toggle_motion":
		currentMovingState := xm.startMoving.Load()
		xm.startMoving.Store(!currentMovingState)
	}
}

func (xm *XWindowManager) handleConfigureRequest(evWindow WindowID, data *ConfigureRequestEvent) {
	fmt.Println("XWM: Handling ConfigureRequest for window \"", evWindow, "\"")

	var changes WindowChanges
	changes.X = data.X
	changes.Y = data.Y
	changes.Width = data.Width
	changes.Height = data.Height
	changes.BorderWidth = data.BorderWidth
	changes.Sibling = data.Above
	changes.StackMode = data.Detail

	clientContext, err := xm.clientManager.FindByClient(evWindow)

	if err != nil || clientContext == nil || clientContext.Frame == NoWindow {
		xm.backend.ConfigureWindow(evWindow, data.ValueMask, &changes)
		xm.backend.Flush()
		return
	}

	// Keep the geometry used by animation in sync with accepted requests.
	if data.ValueMask&CWX != 0 {
		clientContext.Geom.X = data.X
	}
	if data.ValueMask&CWY != 0 {
		clientContext.Geom.Y = data.Y
	}
	if data.ValueMask&CWWidth != 0 {
		clientContext.Geom.Width = data.Width
	}
	if data.ValueMask&CWHeight != 0 {
		clientContext.Geom.Height = data.Height
	}
	changes.Width = clientContext.Geom.Width + TitlebarWidth

	xm.backend.ConfigureWindow(clientContext.Frame, data.ValueMask, &changes)

	xm.backend.MoveResizeWindow(
		clientContext.Window,
		TitlebarWidth,
		0,
		uint(clientContext.Geom.Width),
		uint(clientContext.Geom.Height),
	)

	xm.backend.Flush()
}

func (xm *XWindowManager) handleExpose(evWindow WindowID, data *ExposeEvent) {
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
	height := uint(rect.Height)
	gc := xm.backend.CreateGC(evWindow)
	defer xm.backend.FreeGC(gc)

	// drawing rectable
	// using color 0x3b82f6
	const backgroundColor uint64 = 0x3B82F6
	xm.backend.SetForeground(gc, backgroundColor)
	xm.backend.FillRectangle(evWindow, gc, 0, 0, TitlebarWidth, height)

	// draw a text
	xm.backend.SetForeground(gc, 0xFFFFFF)
	startY := 20
	charSpacing := 12
	textX := 6

	runes := []rune(clientWin.Title)

	for i, r := range runes {
		currentY := startY + i*charSpacing

		if currentY > int(height)-10 {
			break
		}

		xm.backend.DrawString(evWindow, gc, textX, currentY, string(r), 1)
	}

}

func (xm *XWindowManager) handleButtonPress(evWindow WindowID, data *ButtonEvent) {
	fmt.Println("XWM: Handling ButtonPress for window \"", evWindow, "\"")
	if clientContext, err := xm.clientManager.FindByFrame(evWindow); err == nil && clientContext != nil {
		if data.X < 0 {
			xm.backend.ReplayPointer(data.Time)
			xm.backend.Flush()
			return
		}

		xm.FocusClient(clientContext, data.Time)
		if data.X >= TitlebarWidth {
			xm.backend.ReplayPointer(data.Time)
			xm.backend.Flush()
			return
		}

		clientContext.Motion.LastTime = data.Time
		clientContext.Motion.VelocityX = 0
		clientContext.Motion.VelocityY = 0
		clientContext.Motion.StartMouseX = data.XRoot
		clientContext.Motion.StartMouseY = data.YRoot

		clientContext.Motion.StartAttr = xm.backend.GetWindowAttributes(clientContext.Frame)
		clientContext.Geom.X = clientContext.Motion.StartAttr.X
		clientContext.Geom.Y = clientContext.Motion.StartAttr.Y
		clientContext.Motion.LastX, clientContext.Motion.LastY = clientContext.Geom.X, clientContext.Geom.Y

		xm.clientManager.SetMouseMovedClient(clientContext)

		// The active drag grab replaces the passive grab and thaws the pointer.
		xm.backend.GrabPointer(evWindow)
	}
}
func (xm *XWindowManager) handleButtonRelease(evWindow WindowID, data *ButtonEvent) {
	fmt.Println("XWM: Handling ButtonRelease for window \"", evWindow, "\"")
	if clientContext := xm.clientManager.GetMouseMovedClient(); clientContext != nil {
		lastX, lastY := xm.backend.TranslateCoordinates(clientContext.Frame, xm.rootWindow, 0, 0)
		if data.Time != 0 && clientContext.Motion.LastTime != 0 && data.Time-clientContext.Motion.LastTime > 100 {
			clientContext.Motion.VelocityX, clientContext.Motion.VelocityY = 0, 0
		}
		xm.backend.UngrabPointer()
		// update geometry - setting new X,Y positions
		clientContext.Geom.X = lastX
		clientContext.Geom.Y = lastY
		xm.clientManager.SetMouseMovedClient(nil)
		xm.backend.Flush()
	}
}
func (xm *XWindowManager) handleMotionEvent(evWindow WindowID, data *MotionEvent) {
	fmt.Println("XWM: Handling MotionEvent for window \"", evWindow, "\"")
	if clientContext := xm.clientManager.GetMouseMovedClient(); clientContext != nil {
		deltaX := data.XRoot - clientContext.Motion.StartMouseX
		deltaY := data.YRoot - clientContext.Motion.StartMouseY

		newX := clientContext.Motion.StartAttr.X + deltaX
		newY := clientContext.Motion.StartAttr.Y + deltaY
		// Velocity is pixels per animation tick, sampled using X server time.
		elapsed := uint32(movementInterval.Milliseconds())
		if data.Time != 0 && clientContext.Motion.LastTime != 0 {
			elapsed = data.Time - clientContext.Motion.LastTime
		}
		if elapsed > 0 {
			clientContext.Motion.VelocityX = dragVelocity(newX-clientContext.Motion.LastX, elapsed)
			clientContext.Motion.VelocityY = dragVelocity(newY-clientContext.Motion.LastY, elapsed)
			clientContext.Motion.LastTime = data.Time
			clientContext.Motion.LastX, clientContext.Motion.LastY = newX, newY
		}
		clientContext.Geom.X, clientContext.Geom.Y = newX, newY
		xm.backend.MoveWindow(clientContext.Frame, newX, newY)
		xm.backend.Flush()
	}
}

// Transfer a quarter of the mouse speed, capped at 12 pixels per tick.
const throwSpeedFactor = 0.25

func dragVelocity(delta int, elapsed uint32) int {
	velocity := float64(delta) * float64(movementInterval.Milliseconds()) / float64(elapsed) * throwSpeedFactor
	return int(math.Round(max(-12, min(12, velocity))))
}
