package xwindowmgr

/*
#include <X11/Xlib.h>
#include <X11/keysym.h>

*/
import "C"
import (
	"context"
	"time"
)

type WMEvent struct {
	Type   int
	Window C.Window
	Data   interface{}
}

func (xm *XWindowManager) EventLoop(ctx context.Context, eventChan chan<- WMEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if xm.backend.Pending() > 0 {
				ev := xm.backend.NextEvent()

				evType := unionToInt(ev)
				wmEv := xm.parseXEvent(evType, ev)
				eventChan <- wmEv
			} else {
				time.Sleep(15 * time.Millisecond)
			}

		}
	}
}
