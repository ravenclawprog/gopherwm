package xwindowmgr

import "time"

const movementInterval = 20 * time.Millisecond

// advanceAxis confines the frame to the screen and reflects only outward motion.
// Oversized frames are pinned to the origin on this axis.
func advanceAxis(position, velocity, extent, screen int) (int, int) {
	limit := screen - extent
	if limit <= 0 {
		return 0, 0
	}
	next := position + velocity
	if next <= 0 {
		next = 0
		if velocity < 0 {
			velocity = -velocity
		}
	} else if next >= limit {
		next = limit
		if velocity > 0 {
			velocity = -velocity
		}
	}
	return next, velocity
}

func (xm *XWindowManager) moveWindows() {
	dragged := xm.clientManager.GetMouseMovedClient()
	moved := false
	for _, client := range xm.clientManager.GetAllClients() {
		if client == dragged {
			continue
		}
		rect, motion := &client.Geom, &client.Motion
		rect.X, motion.VelocityX = advanceAxis(rect.X, motion.VelocityX, rect.Width+TitlebarWidth, xm.rootGeom.Width)
		rect.Y, motion.VelocityY = advanceAxis(rect.Y, motion.VelocityY, rect.Height, xm.rootGeom.Height)
		xm.backend.MoveWindow(client.Frame, rect.X, rect.Y)
		moved = true
	}
	if moved {
		xm.backend.Flush()
	}
}
