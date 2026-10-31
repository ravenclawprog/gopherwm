package xwindowmgr

import (
	"context"
	"time"
)

func (xm *XWindowManager) MovingWindowsLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(20 * time.Millisecond):
			if xm.startMoving.Load() {
				xm.moveWindows()
			}
		}
	}
}

func (xm *XWindowManager) moveWindows() {
	for _, client := range xm.clientManager.GetAllClients() {
		if mouseMoveClient := xm.clientManager.GetMouseMovedClient(); mouseMoveClient == nil || client.Window != mouseMoveClient.Window {
			var rect *WindowRect = &client.Geom
			var move *MotionData = &client.Motion
			rect.X = rect.X + move.VelocityX
			rect.Y = rect.Y + move.VelocityY

			if rect.X+rect.Width > xm.rootGeom.Width || (rect.X < 0) {
				move.VelocityX *= -1
			}
			if rect.Y+rect.Height > xm.rootGeom.Height || (rect.Y < 0) {
				move.VelocityY *= -1
			}

			xm.backend.MoveWindow(client.Frame, rect.X, rect.Y)
			xm.backend.Flush()
		}
	}
}
