package xwindowmgr

import (
	"context"
	"fmt"
)

// dispatchEvent checks the payload before invoking the matching window handler.
func dispatchEvent[T any](event WMEvent, handler func(WindowID, *T)) error {
	data, ok := event.Data.(T)
	if !ok {
		return fmt.Errorf("event %d: expected payload %T, got %T", event.Type, *new(T), event.Data)
	}
	handler(event.Window, &data)
	return nil
}

func (xm *XWindowManager) handleWMEvent(event WMEvent) error {
	switch event.Type {
	case MapRequest:
		return dispatchEvent(event, xm.handleMapRequest)
	case UnmapNotify:
		return dispatchEvent(event, xm.handleUnmapNotify)
	case KeyPress:
		return dispatchEvent(event, xm.handleKeyPress)
	case ConfigureRequest:
		return dispatchEvent(event, xm.handleConfigureRequest)
	case Expose:
		return dispatchEvent(event, xm.handleExpose)
	case ButtonPress:
		return dispatchEvent(event, xm.handleButtonPress)
	case ButtonRelease:
		return dispatchEvent(event, xm.handleButtonRelease)
	case MotionNotify:
		return dispatchEvent(event, xm.handleMotionEvent)
	case DestroyNotify:
		return dispatchEvent(event, xm.handleDestroyNotify)
	default:
		return fmt.Errorf("unhandled event type: %d", event.Type)
	}
}

func (xm *XWindowManager) HandleWMEvent(ctx context.Context) {

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-xm.eventChan:
			if !ok {
				return
			}
			if err := xm.handleWMEvent(event); err != nil {
				fmt.Println("XWM:", err)
			}
		}
	}
}
